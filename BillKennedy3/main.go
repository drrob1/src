package main
import "fmt"

func main() {
    friends := []string{"Apple", "Banana", "Charley", "Delta", "Echo"}
    for _, v := range friends {
        friends = friends[:2]
        fmt.Printf(" v[%s]\n", v)
    }
    fmt.Println("\n\n\n")

    friends = []string{"Apple", "Banana", "Charley", "Delta", "Echo"}
    for i := range friends {
        friends = friends[:2]
        fmt.Printf(" v[%s]\n", friends[i])
    }
}
