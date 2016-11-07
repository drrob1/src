<*DEFINE (ConsoleMode,TRUE)*>
<*DEFINE (TokenPtr,TRUE)*>

MODULE CompareSHA1;
(*
  REVISION HISTORY
  ----------------
   4 Apr 13 -- First version of module, using computesha1 as a template.

*)
  IMPORT Terminal,MiscM2,SYSTEM,FileFunc;
  FROM SYSTEM IMPORT ADDRESS, ADR;
  FROM FilePicker IMPORT FileNamePicker;
  FROM Storage IMPORT ALLOCATE, DEALLOCATE;
  FROM RealStr IMPORT StrToReal, RealToFloat, RealToEng, RealToFixed, RealToStr;
  IMPORT RConversions, LongStr, LongConv, WholeStr;
  FROM FileFunc IMPORT EOL, FileSpecString, NameString, FileAttributes, FileAttributeSet,
    SearchEntry, FileNameParts (*drive path name extension*), FileTypes, DeviceTypes,
    AccessModes, FileUseInfo, FileUseInfoSet, CommonFileErrors, File, InvalidHandle,
    MustHaveNormalFile, MustHaveDirectory, MustHaveNothing, AllAttributes, StdAttributes,
    AddArchive, AddReadOnly, AddHidden, AddSystem, AddCompressed, AddTemporary,
    AddEncrypted, AddOffline, AddAlias, AddNormalFile, AddDirectory, OpenFile,
    OpenFileEx, CreateFile, CreateFileEx, GetTempFileDirectory, MakeTempFileName,
    CreateTempFile, CreateTempFileEx, OpenCreateFile, OpenCreateFileEx, FakeFileOpen,
    CloseFile, FileType, SetFileBuffer, RemoveFileBuffer, FlushBuffers, ReadBlock,
    WriteBlock, (* ReadChar, WriteChar,*) PeekChar, ReadLine, WriteLine, LockFileRegion,
    UnlockFileRegion, SetFilePos, GetFilePos, MoveFilePos, TruncateFile, FileLength,
    GetFileSizes, TranslateFileError, GetFileAttr, SetFileAttr, GetFileDateTime,
    SetFileDateTime, RenameFile, DeleteFile,
    FileExists, CopyFile, SetHandleCount, GetNextDir, ParseFileName, ParseFileNameEx,
    AssembleParts, ConstructFileName, ConstructFileNameEx, FindInPathList,
    FindInOSPathList, ExpandFileSpec, FindFirst, FindNext, FindClose,
    MakeDir, CreateDirTree, DeleteDir, DirExists, RenameDir, GetDefaultPath,
    SetDefaultPath, GetDeviceFreeSpace, GetDeviceFreeSpaceEx, GetDeviceType;
  IMPORT BasicDialogs;
  FROM BasicDialogs IMPORT MessageTypes;
  IMPORT Strings,MemUtils,ASCII;
  FROM Environment IMPORT GetCommandLine;
  FROM Strings IMPORT
    Append, Equal, Delete, Concat, Capitalize;
  FROM ExStrings IMPORT
    AppendChar, EqualI, AssignNullTerm, Lowercase;
  FROM UTILLIB IMPORT NULL,CR,BUFSIZ,CTRLCOD,STRTYP,STR10TYP,BUFTYP,MAXCARDFNT,COPYLEFT,COPYRIGHT,FILLCHAR,SCANFWD,SCANBACK,
    SubStrCMPFNT,STRLENFNT,STRCMPFNT,LCHINBUFFNT,MRGBUFS,TRIMFNT,TRIM,RMVCHR,APPENDA2B,CONCATAB2C,INSERTAin2B,ASSIGN2BUF;
  FROM MyFIO2 IMPORT EOFMARKER,DRIVESEP,SUBDIRSEP,EXTRACTDRVPTH,MYFILTYP,
    IOSTATE,FRESET,FPURGE,FOPEN,FCLOSE,FREAD,FRDTXLN,FWRTX,FWRTXLN,RETBLKBUF,
    FWRSTR,FWRLN,FAPPEND,COPYDPIF,GETFNM;

  FROM SHA1 IMPORT SHA1, Create, Destroy, Reset, HashBytes, GetString;

%IF ConsoleMode %THEN
    IMPORT MiscStdInOut, SIOResult;
    FROM MiscStdInOut IMPORT WriteCard, WriteLongCard, CLS, WriteString, WriteLn, PressAnyKey, Error, WriteInt,
      WriteReal, WriteLongReal, WriteChar, ReadChar, ReadString, SkipLine, ReadCard, ReadLongReal;
%ELSE
    IMPORT MiscM2;
    FROM MiscM2 IMPORT WriteCard, WriteLongCard, CLS, WriteString, WriteLn, PressAnyKey, Error, WriteInt,
      WriteReal, WriteLongReal, WriteChar, ReadChar, Read, ReadString, ReadCard, ReadLongReal;
%END
<* IF TokenPtr THEN *>
    FROM TOKENPTR IMPORT FSATYP,DELIMCH,DELIMSTATE,INI1TKN,TKNPTRTYP,
      INI3TKN,GETCHR,UNGETCHR,GETTKN,GETTKNSTR,GETTKNEOL,UNGETTKN,GETTKNREAL;
<* ELSE *>
    FROM TKNRTNS IMPORT FSATYP,CHARSETTYP,DELIMCH,DELIMSTATE,INI1TKN,
      INI3TKN,GETCHR,UNGETCHR,GETTKN,NEWDELIMSET,NEWOPSET,NEWDGTSET,
      GETTKNSTR,GETTKNEOL,UNGETTKN,GETTKNREAL;
<* END *>

  CONST
    K = 1024;

  TYPE
    HashType = (md5hash, sha1hash, sha256hash, sha384hash, sha512hash);
                                                (*  FileOrHashType = (filetoken, hashtoken); Don't think I need this now *)

  VAR
    C,IDX,PTR,c,RETCOD                             : CARDINAL;
    CH,ch                                          : CHAR;
    FLAG,FLAG2,FLAG3,ignoreboolean,EOFFLG,stop     : BOOLEAN;
    I,J,i,j                                        : INTEGER;
    TargetFile                                     : File;
    HashesList                                     : MYFILTYP;
    TKNSTATE                                       : FSATYP;
    PROMPT,NAMDFT,TYPDFT,INFNAM,OUTFNAM,TargetFileNameBuf,
    TMPBUF,NUMBUF,DRVPATH,INBUF,TOKEN              : BUFTYP;
    InNameStr,OutName,OldInName                    : NameString;
    innameparts,outnameparts                       : FileNameParts;
    entry                                          : SearchEntry;
    inputline,OpenFileName                         : ARRAY [0..255] OF CHAR;
    HashValueInList,HashStrComputed,TargetFileName : STRTYP;
    WhichHash                                      : HashType;
                                                      (*  isfileorhash                                 : FileOrHashType; *)
    InBuf                                          : ARRAY [1..100*K] OF CHAR;
    a                                              : ADDRESS;
    filesize                                       : LONGCARD;
    hash                                           : SHA1;
<* IF TokenPtr Then *>
    tpv                                            : TKNPTRTYP;
<* END *>


(*
These file routines will read byte by byte without any buffering or making sure an entire
line is read in.  So these are simpler than the routines
*)

PROCEDURE fopen(VAR INOUT F:File; FILNAM:ARRAY OF CHAR; RDWRSTATE:IOSTATE);
(*
************************ fopen **************************************
File Open.
RDWRSTATE IS EITHER RD FOR OPEN A FILE FOR READING, OR WR FOR OPEN A FILE FOR WRITING.
I wrote years before.
*)

  VAR
    I,RETCOD : CARDINAL;
    EOFFLG   : BOOLEAN;
    fileError: CommonFileErrors;
    filelength : CARDINAL32;

  BEGIN
    CASE RDWRSTATE OF
      RD : OpenFile(F,FILNAM,ReadOnlyDenyWrite);
    | WR : CreateFile(F,FILNAM);  (*This proc truncates file before creation*)
    | APND : OpenCreateFile(F,FILNAM,ReadWriteDenyWrite);
    END(*CASE*);
    IF F.status <> 0 THEN
      WriteString(' Error in opening/creating file ');
      WriteString(FILNAM);
      WriteString('--');
      CASE TranslateFileError(F) OF
        FileErrFileNotFound : WriteString('File not found.');
      | FileErrDiskFull : WriteString('Disk Full');
      ELSE
        WriteString('Nonspecific error occured.');
      END(*CASE*);
      WriteLn;
      WriteString(' Program Terminated.');
      WriteLn;
      HALT;
    END(*IF F.status*);

    IF RDWRSTATE = APND THEN
      filelength := FileLength(F);
      MoveFilePos(F,filelength);
    END(*IF APND*);

  END fopen;


PROCEDURE fclose(VAR INOUT F:File);
  BEGIN
    CloseFile(F);
  END fclose;



(* ************************* MAIN ***************************************************************)


BEGIN

<* IF NOT ConsoleMode THEN *>
  Terminal.Reset;
<* END *>

  FileNamePicker(InNameStr);
  IF LENGTH(InNameStr) < 1 THEN
    c := 0;
    FLAG := BasicDialogs.PromptOpenFile(InNameStr,'',c,'','','Open input text file',FALSE);
    IF NOT FLAG THEN
      Error('Could not open file.  Does it exist?');
      HALT;
    END; (* if not flag for BasicDialogs promptopenfile *)
  END;  (* if length(innamestr) from filepicker *)
  ParseFileName(InNameStr, innameparts);
  Lowercase(innameparts.extension);
  IF STRCMPFNT(innameparts.extension,'.md5') = 0 THEN
    WhichHash := md5hash;
  ELSIF STRCMPFNT(innameparts.extension,'.sha1') = 0 THEN
    WhichHash := sha1hash;
  ELSIF STRCMPFNT(innameparts.extension,'.sha256') = 0 THEN
    WhichHash := sha256hash;
  ELSIF STRCMPFNT(innameparts.extension,'.sha384') = 0 THEN
    WhichHash := sha384hash;
  ELSIF STRCMPFNT(innameparts.extension,'.sha512') = 0 THEN
    WhichHash := sha512hash;
  ELSE
    WriteString(' Not a recognized hash extension.  Will assume sha1.  For now.');
    WriteLn;
    WhichHash := sha1hash;
  END; (* IF HashType *)
  WriteString(' Testing determining hash type by file extension.');
  WriteLn;
  WriteString(' HashType = md5, sha1, sha256, sha384, sha512.  WhichHash = ');
  WriteCard(ORD(WhichHash));
  WriteLn;

(* Have HashesList name in InName.  Must parse it into the 2 parts *)
  ASSIGN2BUF(InNameStr,INFNAM);
  FOPEN(HashesList,INFNAM,RD);

  LOOP (* to read multiple lines *)
  stop := FALSE;
  filesize := 0;
    FRDTXLN(HashesList,INBUF,0);
    IF HashesList.FILE.eof THEN EXIT END(*IF*);
    IF (INBUF.CHARS[1] = ';') OR (INBUF.CHARS[1] = '#') THEN CONTINUE END; (* allow comments *)
    INI1TKN(tpv,INBUF);
    GETTKNSTR(tpv,TOKEN,I,RETCOD);
    IF RETCOD > 0 THEN
      Error(' Error while tokenizing line in the file.  Skipping');
      CONTINUE;
    END; (* if retcod >0 *)
    IF LCHINBUFFNT(TOKEN,'.') > 0 THEN (* have filename first on line *)
      AssignNullTerm(TOKEN.CHARS,TargetFileName);
      GETTKNSTR(tpv,TOKEN,I,RETCOD);  (* Get stored hash *)
      IF RETCOD > 0 THEN
        Error(' Error while tokenizing line in the file.  Skipping');
        CONTINUE;
      END; (* if retcod >0 *)
      HashValueInList := TOKEN.CHARS;
    ELSE  (* have hash first on line *)
      HashValueInList := TOKEN.CHARS;
      GETTKNSTR(tpv,TOKEN,I,RETCOD);
      IF RETCOD > 0 THEN
        Error(' Error while tokenizing line in the file.  Skipping');
        CONTINUE;
      END; (* if retcod >0 *)
      TargetFileName := TOKEN.CHARS;
    END; (* if have filename first or hash value first *)
    GETTKNSTR(tpv,TOKEN,I,RETCOD);  (* nothing left on line.  This call is to get an EOL condition and DISPOSE the tpv pointer. *)
(*
  now to compute the hash, compare them, and output results
*)
    (* Create Hash Section *)
    fopen(TargetFile,TargetFileName,RD);
    a := ADR(InBuf);
    hash := Create();
    Reset(hash);
    REPEAT
      ReadBlock(TargetFile,a,SIZE(InBuf));
      IF (TargetFile.status > 0) AND NOT TargetFile.eof THEN
        WriteString(' Error from file ReadBlock.  FileSize is .');
        WriteLongCard(filesize);
        WriteLn;
        WriteString(' Halting ');
        WriteLn;
        fclose(TargetFile);
        HALT;
      END; (* if Targetfile.status error *)
      INC(filesize,TargetFile.count);
      HashBytes(hash,a,TargetFile.count);
    UNTIL TargetFile.eof;

    GetString(hash,HashStrComputed);
    WriteString(' Filename = ');
    WriteString(InNameStr);
    WriteString(', FileSize = ');
    WriteLongCard(filesize);
    WriteLn;
    WriteString(' SHA1 computed hash string is : ');
    WriteString(HashStrComputed);
    WriteLn;
    IF STRCMPFNT(HashValueInList,HashStrComputed) = 0 THEN
      WriteString(" Matched.");
    ELSE
      WriteString(' Not matched.');
    END; (* if hashes *)
    WriteString(" Downloaded hash is ");
    WriteString(HashValueInList);
    WriteLn;
  END(* outer LOOP to read multiple lines*);
  fclose(TargetFile);
  FCLOSE(HashesList);

<*IF NOT ConsoleMode THEN  *>
  PressAnyKey;
<* END *>

END CompareSHA1.
(*
PROCEDURE RenameFile(fromFile, toFile : ARRAY OF CHAR) : BOOLEAN;
PROCEDURE DeleteFile(name : ARRAY OF CHAR) : BOOLEAN;
PROCEDURE FileExists(name : ARRAY OF CHAR) : BOOLEAN;
PROCEDURE CopyFile(source, dest : ARRAY OF CHAR) : BOOLEAN;
Type
  File = RECORD   And I will only list fields of importance to me.
    status : CARDINAL;  error code of last operation.  0 if successful.
    count  : CARDINAL;  used by the read and write procs
    eof  : BOOLEAN;
  END record
FileFunc PROCEDURE ReadBlock(VAR INOUT f: File; buf : ADDRESS; size : CARDINAL);  File.count contains actual amount read.  File.success =0 if successful
SHA1 PROCEDURE Create ():SHA1;
SHA1 PROCEDURE Reset(hash:SHA1);
SHA1 PROCEDURE HashBytes(hash:SHA1; data:ADDRESS; amount:CARDINAL);
SHA1 PROCEDURE GetString(hash:SHA1; VAR OUT str:ARRAY OF CHAR);  str is 40 characters long.


*)
(*
I want to do the following
There are 2 read files, TargetFile and HashesFile.
  TargetFile is the file to be read and computed.
  HashesFile lists the name of this file and the published hash to be compared w/ the computed hash.  Since the published hash is a string,
   this is what will be compared.
So I will need to import UTILLIB and look more closely in the routines in Strings and ExStrings.

Hashesfile will have 2 string tokens per line delimited by spaces.  Since the hash can begin w/ a digit, these TKNSTATE is ALLELSE and I
probably need to use GETTKNSTR or whatever I called it.
One string is a filename and will have a '.', and the other will not.  I want to be able to parse lines either way, based on whether
there is a '.' or not.  If not, it is a HAsH string.  If yes, it is a filename to be hashed.  So only the Hashesfile has to be specified
by the user.  The rest is automatic.  And there may be more than 1 line to be processed.

Fetch HashesFile.  Parse out its extension.  Do nothing w/ that now, but in future will use that to determine which hash functions get called.
  Since this is my eventual plan, I do need to break out the computation of hashes into separate procedures, all of which use TargetFile which must
  already be opened.  I see this as TYPE HashType = (md5, sha1, sha256, sha384, sha512);  These names won't conflict w/ the opaque data type exported by
  these modules, as long as I keep them all lowercase.  Or make them (md5hash, sha1hash, sha256hash, sha384hash, sha512hash);
Get 1st tokenstr.  If has '.' it's a filename, else it's a hash string.  TYPE FileOrHash = (filetoken, hashtoken);
Get 2nd token.  It's whatever the 1st was not.

Compute Hash on TargetFile
Display filename, size and hash value, and whether there is a MATCH or NO MATCH.


Type
  HashType = (md5hash, sha1hash, sha256hash, sha384hash, sha512hash);
  FileOrHashType = (filetoken, hashtoken);
VAR
  TargetFile, HashesList : File;
  isfileorhash : FileOrHashType;   /* this will hold the state of the 1st token on the line, so decisions can be made already knowing what
                                    the 2nd token must be. */
  WhichHash : HashType;
  hashvalueinlist : STRTYP;

begin
  get HashesList from neat little trick I wrote for first using filenamepicker and then a pop-up.
  get filenameparts to analyze the extension.  set WhichHash
  REPEAT
  get 1st token.  If has '.' in it, set isfileorhash to be filetoken, else hashtoken.
  If a filetoken, it is a TargetFile, else it is a hashvalueinlist string value.
  get 2nd token, knowing it must be what the 1st one was not.

  Now have both TargetFile and hashvalueinlist.  Compute the hash and compare to hashvalueinlist.

  Write out filename, size, computed hash and matched or not match.  If not matched, writeout hashvalueinlist.
  UNTIL HashesList.eof

Thats all I can think of now.




*)
