package test

import (
	"context"
	"fmt"
)

type HelloWorld struct {
}

func (h *HelloWorld) Hello(ctx context.Context, name string) (string, error) {

	fmt.Println(ctx.Value("User-Agent"))
	return "hello," + name, nil
}
