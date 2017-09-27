go client for <http://www.deathbycaptcha.com/>

##### install 

go get github.com/zhanghjster/go-deathbycaptcher

##### usage

```go
import (
	"os"
	"io/ioutil"
	"log"
)
func main() {
	if len(os.Args) < 1 {
		log.Fatal("image file not defined")
	}

  	// read the captcha image
	f, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	var c = &Client{
		Username: "your_deathbycaptcha_username",
		Password: "your_deathbycaptcha_password",
	}

	code, err := c.Decode(f)
	if err != nil {
		log.Fatal(err)
	}

	println("code is ", code)
}

$ go run main.go captcher.png 
```

**set system enviroment DBC_MODE to "dev" to show the log**







