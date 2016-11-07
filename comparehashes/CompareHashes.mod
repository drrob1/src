<*DEFINE (ConsoleMode,TRUE)*>
<*DEFINE (TokenPtr,TRUE)*>

MODULE CompareHashes;
(*
  REVISION HISTORY
  ----------------
   6 Apr 13 -- First version of module, using ComputeSHA1 as a template.
  23 Apr 13 -- Fixed problem of a single line in the hashes file, that does not contain an EOL character, causes
                an immediate return without processing of the characters just read in.
  24 Apr 13 -- Added output of which file either matches or does not match.

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

  IMPORT MD5, SHA1, SHA256, SHA384, SHA512;

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
    Str10ArrayType = ARRAY [0..ORD(sha512hash)] OF STR10TYP;

  CONST
    HashName = Str10ArrayType {'md5','sha1','sha256','sha384','sha512'};

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
    InBuf                                          : ARRAY [1..100*K] OF CHAR;
    a                                              : ADDRESS;
    filesize                                       : LONGCARD;
    hash1                                          : SHA1.SHA1;
    hash5                                          : MD5.MD5;
    hash256                                        : SHA256.SHA256;
    hash384                                        : SHA384.SHA384;
    hash512                                        : SHA512.SHA512;

<* IF TokenPtr Then *>
    tpv                                            : TKNPTRTYP = NIL;
<* END *>


(*
These file routines will read byte by byte without any buffering or making sure an entire
line is read in.  So these are simpler than the routines in MyFIO(2)
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

  Terminal.Reset;  (* Note that FileNamePicker uses terminal mode, even the rest of the I/O is console mode *)

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
(*
  WriteString(' Testing determining hash type by file extension.');
  WriteLn;
  WriteString(' HashType = md5, sha1, sha256, sha384, sha512.  WhichHash = ');
  WriteCard(ORD(WhichHash));
  WriteLn;
*)

(* Have HashesList name in InName.  Must parse it into the 2 parts *)
  ASSIGN2BUF(InNameStr,INFNAM);
  FOPEN(HashesList,INFNAM,RD);
  WriteString(' File containing hashes is ');
  WriteString(InNameStr);
  WriteLn;

  REPEAT (* to read multiple lines *)
    stop := FALSE;
    filesize := 0;
    FRDTXLN(HashesList,INBUF,0);
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
    CASE WhichHash OF
      md5hash :
           hash5 := MD5.Create();
           MD5.Reset(hash5);
    | sha1hash :
           hash1 := SHA1.Create();
           SHA1.Reset(hash1);
    | sha256hash :
           hash256 := SHA256.Create();
           SHA256.Reset(hash256);
    | sha384hash :
           hash384 := SHA384.Create();
           SHA384.Reset(hash384);
    | sha512hash :
           hash512 := SHA512.Create();
           SHA512.Reset(hash512);
      ELSE
    END; (* case on WhichHash *)
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
      CASE WhichHash OF
        md5hash : MD5.HashBytes(hash5,a,TargetFile.count);
      | sha1hash : SHA1.HashBytes(hash1,a,TargetFile.count);
      | sha256hash : SHA256.HashBytes(hash256,a,TargetFile.count);
      | sha384hash : SHA384.HashBytes(hash384,a,TargetFile.count);
      | sha512hash : SHA512.HashBytes(hash512,a,TargetFile.count);
        ELSE
      END; (* case on WhichHash *)
    UNTIL TargetFile.eof;

    CASE WhichHash OF
      md5hash : MD5.GetString(hash5,HashStrComputed);
    | sha1hash : SHA1.GetString(hash1,HashStrComputed);
    | sha256hash : SHA256.GetString(hash256,HashStrComputed);
    | sha384hash : SHA384.GetString(hash384,HashStrComputed);
    | sha512hash : SHA512.GetString(hash512,HashStrComputed);
      ELSE
    END; (* case on WhichHash *)
    WriteString(' Filename  = ');
    WriteString(TargetFileName);
    WriteString(', FileSize = ');
    WriteLongCard(filesize);
    WriteString(', ');
    WriteString(HashName[ORD(WhichHash)]);
    WriteString(' computed hash string, followed by hash string in the file are : ');
    WriteLn;
    WriteString(HashStrComputed);
    WriteLn;
    WriteString(HashValueInList);
    WriteLn;
    IF STRCMPFNT(HashValueInList,HashStrComputed) = 0 THEN
      WriteString(" Matched.");
    ELSE
      WriteString(' Not matched.');
      IF SubStrCMPFNT(HashStrComputed,HashValueInList) THEN
        WriteString('   However computed hash is a substring of the hash string in the file.');
      END; (* if substring *)
    END; (* if hashes *)
    WriteLn;
    WriteLn;
    WriteLn;
  UNTIL HashesList.FILE.eof; (* outer LOOP to read multiple lines*);
  fclose(TargetFile);
  FCLOSE(HashesList);
  WriteLn;
<*IF NOT ConsoleMode THEN  *>
  PressAnyKey;
<* END *>

END CompareHashes.
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
