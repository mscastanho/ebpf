package examples

// DocRlimit {
import "github.com/mscastanho/ebpf/rlimit"

func main() {
	if err := rlimit.RemoveMemlock(); err != nil {
		panic(err)
	}
}

// }
