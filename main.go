// create by chencanhua in 2023/5/7
package main

import "fmt"

func main() {
	number2 := []int{1, 2, 3, 4, 5, 6}
	maxIndex2 := len(number2) - 1
	for i, e := range number2 {
		if i == maxIndex2 {
			number2[0] += e
		} else {
			number2[i+1] += e
		}
	}
	fmt.Println(number2)
}
