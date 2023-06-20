package test

import (
	"context"
	"log"
	"strings"
	"testing"

	"github.com/zhouyangtingwen/dify-sdk-go"
)

func TestApi3(t *testing.T) {
	var (
		host = ""
		apiSecretKey = ""
	)

	var c = &dify.ClientConfig{
		Host: host,
		ApiSecretKey: apiSecretKey,
	}
	var client1 = dify.NewClientWithConfig(c)

	var client2 = dify.NewClient(host, apiSecretKey)

	t.Log(client1.GetHost() == client2.GetHost())
	t.Log(client1.GetApiSecretKey() == client2.GetApiSecretKey())

	var ctx, _ = context.WithCancel(context.Background())

	var (
		ch = make(chan dify.ChatMessageStreamChannelResponse)
		err error
	)

	ch, err = client1.Api().ChatMessagesStream(ctx, &dify.ChatMessageRequest{
		Query: "你是谁?",
		User: "jiuquan AI",
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	var (
		strBuilder strings.Builder
		cId string
	)
	for {
		select {
		case <-ctx.Done():
			t.Log("ctx.Done", strBuilder.String())
			return
		case r, isOpen := <-ch:
			if !isOpen {
				goto M
			}
			log.Println("Answer2", r.Answer)
			strBuilder.WriteString(r.Answer)
			cId = r.ConversationID
		}
	}

	M:
	t.Log(strBuilder.String())
	t.Log(cId)

	// var msg *dify.MessagesResponse
	// if msg, err = client1.Api().Messages(ctx, &dify.MessagesRequest{
	// 	ConversationID: cId,
	// 	User: "jiuquan AI",
	// }); err != nil {
	// 	t.Fatal(err.Error())
	// 	return
	// }
	// t.Log(msg)
}