package main
import "fmt"
import "math"
 
type Shaper interface {
        Area() float32
}
 
type Square struct {
        side float32
}
 
func (sq *Square) Area() float32 {               // this receiver is a pointer type
        return sq.side * sq.side
}
 
type Rectangle struct {
  length, width float32
}
 
func (r Rectangle) Area() float32 {             // this receiver is not a pointer type, which here is called a value type.
  return r.length * r.width
}

type Circle struct {
  radius float32
}

func (c Circle) Area() float32 {                // this is defined later in the chapter.  I moved it here
  return c.radius*c.radius*math.Pi
}
 
func main() {
  r := Rectangle{5, 3}    // Area() of Rectangle needs a value, ie, not a pointer
  q := &Square{5}         // Area() of Square needs a pointer
  c := Circle{5}          // Area() of circle needs a value, ie, not a pointer

  // shapes := []Shaper{Shaper(r), Shaper(q)}
  // or shorter:
  shapes := []Shaper{r, q, c}        // the example includes c, so I figure it used to be defined here but got moved in a revision.
  fmt.Println("Looping through shapes for area ... ")
  for n, _ := range shapes {
     fmt.Println("Shape details: ", shapes[n])
     fmt.Println("Area of this shape is: ", shapes[n].Area())
  }

  fmt.Println();
  fmt.Println();
}
