// Practice file: internal/storage/practice.go
package storage

import (
    "fmt" 
    "math"
)

type vertex struct{
    X, Y float64
}

type myfloat float64 

func (f myfloat) root() float64{
    return float64(math.Sqrt(float64(f)))
}

func (v vertex) Abs() float64{
    return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func BasicTypes() {
    var dbName string = "mydb"
    var tableCount int = 5
    fmt.Printf("Database %s has %d tables\n", dbName, tableCount)
    v := vertex{3,4}
    a := myfloat(2)
    fmt.Println(a.root())
    fmt.Println(v.Abs())
}
