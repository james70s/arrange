package ver

import (
	"fmt"
	"log"
	"strings"
)

var (
	Version = "0.0.1" // Version number
	Build   = ""      // Build number
)

// Logo of application
const banner = `
   
   ▄█   ▄█▄    ▄████████  ▄██████▄   ▄████████    ▄████████ 
  ███ ▄███▀   ███    ███ ███    ███ ███    ███   ███    ███ 
  ███▐██▀     ███    ███ ███    ███ ███    █▀    ███    ███ 
 ▄█████▀      ███    ███ ███    ███ ███          ███    ███ 
▀▀█████▄    ▀███████████ ███    ███ ███        ▀███████████ 
  ███▐██▄     ███    ███ ███    ███ ███    █▄    ███    ███ 
  ███ ▀███▄   ███    ███ ███    ███ ███    ███   ███    ███ 
  ███   ▀█▀   ███    █▀   ▀██████▀  ████████▀    ███    █▀     v%s
  ▀ 

拷贝目录中的图像&视频文件到指定的目录下，并根据文件的修改时间，按年/月/日的方式整理到对应的目录下
© 2022 MetabizAI INC.

____________________________________O/_______
                                    O\
`

func VerString() string {
	return strings.Trim(fmt.Sprintf("%s.%s", Version, Build), ".")
}

func Banner() string {
	return fmt.Sprintf(banner, VerString())
}

func Info() {
	log.Println(Banner())
}
