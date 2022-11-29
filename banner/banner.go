package main

import (
	"fmt"
	"github.com/dimiro1/banner"
	"github.com/mattn/go-colorable"
)

/*
  REVISION HISTORY
  -------- -------
  29 Nov 22 -- exploring how banner works
*/
func main() {
	bannerIsEnabled := true
	bannerIsColorEnabled := true
	numStr := "23.45"
	numStr2 := "-12.98"
	text := "{{" + ".Title \"" + numStr + "\"  \"banner\" 0" + "}}" + numStr2
	banner.InitString(colorable.NewColorableStdout(), bannerIsEnabled, bannerIsColorEnabled, "{{ .Title  \"My Title 1\"  \"banner\" 0  }}")
	fmt.Println()
	banner.InitString(colorable.NewColorableStdout(), bannerIsEnabled, bannerIsColorEnabled, "v={{.GoVersion}}  compiler={{.Compiler}} numCPU={{.NumCPU}}")
	fmt.Println()
	banner.InitString(colorable.NewColorableStdout(), bannerIsEnabled, bannerIsColorEnabled, "{{ .Title  \"My Title null\"  \"\" 0  }}")
	fmt.Println()
	//banner.InitString(colorable.NewColorableStdout(), bannerIsEnabled, bannerIsColorEnabled, "{{ .Title  \"My Title 3\"  \"banner3\" 0  }}")
	//fmt.Println()
	//banner.InitString(colorable.NewColorableStdout(), bannerIsEnabled, bannerIsColorEnabled, "{{ .Title  \"My Title 4\"  \"banner4\" 0  }}")
	//fmt.Println()
	banner.InitString(colorable.NewColorableStdout(), bannerIsEnabled, bannerIsColorEnabled, " some text first:\n{{ .Title  \"My Title\"  \"banner\" 0  }}")
	fmt.Println()
	banner.InitString(colorable.NewColorableStdout(), bannerIsEnabled, bannerIsColorEnabled, text)
	fmt.Println()
}
