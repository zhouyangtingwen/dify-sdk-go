package test

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"

	"github.com/zhouyangtingwen/dify-sdk-go"
)

var (
	host = ""
	apiSecretKey = ""
)

func TestApi3(t *testing.T) {
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
			strBuilder.WriteString(r.Answer)
			cId = r.ConversationID
			log.Println("Answer2", r.Answer, r.ConversationID, cId, r.ID, r.TaskID)
		}
	}

	M:
	t.Log(strBuilder.String())
	t.Log(cId)
}

func TestMessages(t *testing.T) {
	var cId = "ec373942-2d17-4f11-89bb-f9bbf863ebcc"
	var err error
	var ctx, _ = context.WithCancel(context.Background())

	// messages
	var messageReq = &dify.MessagesRequest{
		ConversationID: cId,
		User: "jiuquan AI",
	}

	var client = dify.NewClient(host, apiSecretKey)

	var msg *dify.MessagesResponse
	if msg, err = client.Api().Messages(ctx, messageReq); err != nil {
		t.Fatal(err.Error())
		return
	}
	j, _ := json.Marshal(msg)
	t.Log(string(j))
}

func TestMessagesFeedbacks(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	var ctx, _ = context.WithCancel(context.Background())

	var id = "72d3dc0f-a6d5-4b5e-8510-bec0611a6048"

	var res *dify.MessagesFeedbacksResponse
	if res, err = client.Api().MessagesFeedbacks(ctx, &dify.MessagesFeedbacksRequest{
		MessageID: id,
		Rating: dify.FeedbackLike,
		User: "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestConversations(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	var ctx, _ = context.WithCancel(context.Background())

	var res *dify.ConversationsResponse
	if res, err = client.Api().Conversations(ctx, &dify.ConversationsRequest{
		User: "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestConversationsRename(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	var ctx, _ = context.WithCancel(context.Background())

	var res *dify.ConversationsRenamingResponse
	if res, err = client.Api().ConversationsRenaming(ctx, &dify.ConversationsRenamingRequest{
		ConversationID: "ec373942-2d17-4f11-89bb-f9bbf863ebcc",
		Name: "rename!!!",
		User: "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}

func TestParameters(t *testing.T) {
	var client = dify.NewClient(host, apiSecretKey)
	var err error
	var ctx, _ = context.WithCancel(context.Background())

	var res *dify.ParametersResponse
	if res, err = client.Api().Parameters(ctx, &dify.ParametersRequest{
		User: "jiuquan AI",
	}); err != nil {
		t.Fatal(err.Error())
	}

	j, _ := json.Marshal(res)

	log.Println(string(j))
}