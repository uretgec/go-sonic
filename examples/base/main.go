package main

import (
	"context"
	"fmt"
	"time"

	"github.com/uretgec/go-sonic/sonic"
)

func main() {
	ctx := context.Background()

	// INGEST MODE ----------------------------------------
	sonicSearch := sonic.NewClient(&sonic.Options{
		Addr:         "localhost:1491",
		AuthPassword: "SecretPassword",
		ChannelMode:  sonic.ChannelIngest,
	})

	results, err := sonicSearch.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Results: %v\n", results)

	text := "Bir isim gerekiyor acaba ne olsun"

	var chunks []string

	if sonicSearch.IsPushContentReady(text) {
		chunks = sonicSearch.SplitPushContent(text)
	} else {
		chunks = append(chunks, text)
	}

	for _, text := range chunks {
		results, err := sonicSearch.Push(ctx, "collection", "bucket", "user:1", text, sonic.LangTur).Result()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Results: %v\n", results)
	}

	results, err = sonicSearch.FlushCollection(ctx, "collection").Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Results: %v\n", results)

	results, err = sonicSearch.Pop(ctx, "collection", "bucket", "user:1", text).Result()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Results: %v\n", results)

	// SEARCH MODE ----------------------------------------
	for i := 0; i < 2; i++ {
		time.Sleep(10 * time.Second)

		sonicSearch2 := sonic.NewClient(&sonic.Options{
			Addr:         "localhost:1491",
			AuthPassword: "SecretPassword",
			ChannelMode:  sonic.ChannelSearch,
		})

		results, err = sonicSearch2.Query(ctx, "collection", "bucket", "nededin sen", 10, 0, sonic.LangTur).Slice()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Results: %v\n", results)

		results, err = sonicSearch2.Query(ctx, "collection", "bucket", "ne haber", 10, 0, sonic.LangTur).Result()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Results: %v\n", results)

		results, err = sonicSearch2.Suggest(ctx, "collection", "bucket", "gerek", 10).Slice()
		if err != nil {
			panic(err)
		}

		fmt.Printf("Results: %v\n", results)

		err = sonicSearch2.Quit(ctx).Err()
		if err != nil {
			panic(err)
		}

	}

	// CHANNEL MODE ----------------------------------------
	sonicSearch3 := sonic.NewClient(&sonic.Options{
		Addr:         "localhost:1491",
		AuthPassword: "SecretPassword",
		ChannelMode:  sonic.ChannelSearch,
	})

	results, err = sonicSearch3.Suggest(ctx, "collection", "bucket", "gerek", 10).Slice()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Results: %v\n", results)

	results, err = sonicSearch3.Info(ctx).Slice()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Results: %v\n", results)

	results, err = sonicSearch3.Trigger(ctx, sonic.TriggerActionConsolidate, "").Bool()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Results: %v\n", results)

	err = sonicSearch3.Quit(ctx).Err()
	if err != nil {
		panic(err)
	}

	fmt.Println("bye bye")
}
