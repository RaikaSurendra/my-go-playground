package main

import (
	"fmt"
	"time"

	expr "github.com/expr-lang/expr"
)

type Env struct {
	Posts []Post `expr:"posts"`
}

func (Env) Format(t time.Time) string { // Methods defined on the struct become functions.
	return t.Format(time.RFC822)
}

type Post struct {
	Body         string
	Date         time.Time
	RelatedPosts []Post
}

func main() {
	code := `join(map(posts, Format(.Date) + ": " + .Body + "\n"), "")`

	env := Env{
		Posts: []Post{
			{Body: "Hello, world!", Date: time.Now()},
			{Body: "Another post", Date: time.Now().Add(-24 * time.Hour)},
			{Body: "Could I be wearing any more clothes?", Date: time.Now().Add(-72 * time.Hour)},
		},
	}

	output, err := expr.Eval(code, env) // Pass the struct as an environment.
	if err != nil {
		fmt.Println("Error compiling expression:", err)
		return
	}

	fmt.Print(output)
}
