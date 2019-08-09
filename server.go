package main
import (
	"fmt"
	"github.com/proto-lite/model"
	"net"
)
func main() {
	packets := model.NewPackets()
	fmt.Println("Server is running at localhost:8888")
	conn, err := net.ListenPacket("udp", "localhost:8888")
	if err != nil {
		panic(err)
		}
	defer conn.Close()
	buffer := make([]byte, 1500)
	for {
		length, remoteAddress, err := conn.ReadFrom(buffer)
		packets.AddPacketFromReceiveByte(buffer, remoteAddress)
		if err != nil {
			panic(err)
			}
		fmt.Printf("Received from %v: %v\n",
		remoteAddress, string(buffer[:length]))
		_, err = conn.WriteTo([]byte("Hello from Server"), 
			remoteAddress)
		if err != nil {
			panic(err)
			}
		}
}

/*func receiver(ch <-chan []byte)  {
	for {
		i := <- ch

	}
}*/
