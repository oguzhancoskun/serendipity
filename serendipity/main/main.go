package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type tgMessage struct {
	UpdateID      int `json:"update_id"`
	CallbackQuery struct {
		ID   string `json:"id"`
		From struct {
			ID           int    `json:"id"`
			IsBot        bool   `json:"is_bot"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Username     string `json:"username"`
			LanguageCode string `json:"language_code"`
		} `json:"from"`
		Message struct {
			MessageID int `json:"message_id"`
			From      struct {
				ID        int64  `json:"id"`
				IsBot     bool   `json:"is_bot"`
				FirstName string `json:"first_name"`
				Username  string `json:"username"`
			} `json:"from"`
			Chat struct {
				ID                          int    `json:"id"`
				Title                       string `json:"title"`
				Type                        string `json:"type"`
				AllMembersAreAdministrators bool   `json:"all_members_are_administrators"`
			} `json:"chat"`
			Date     int    `json:"date"`
			Text     string `json:"text"`
			Entities []struct {
				Offset int    `json:"offset"`
				Length int    `json:"length"`
				Type   string `json:"type"`
			} `json:"entities"`
			ReplyMarkup struct {
				InlineKeyboard [][]struct {
					Text         string `json:"text"`
					CallbackData string `json:"callback_data"`
				} `json:"inline_keyboard"`
			} `json:"reply_markup"`
		} `json:"message"`
		ChatInstance string `json:"chat_instance"`
		Data         string `json:"data"`
	} `json:"callback_query"`
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	b := []byte(request.Body)
	var msg = new(tgMessage)
	json.Unmarshal(b, &msg)

	save, _ := json.Marshal(msg.CallbackQuery.Data)

	chatID, _ := json.Marshal(msg.CallbackQuery.Message.Chat.ID)
	messageID, _ := json.Marshal(msg.CallbackQuery.Message.MessageID)

	responseBody := map[string]bool{"ok": true}

	prettyResponseBody, err := json.MarshalIndent(responseBody, "", "  ")
	if err != nil {
		return events.APIGatewayProxyResponse{}, err
	}

	success, _ := updateTable(string(save))

	if success {
		removeKeyboard(string(chatID), string(messageID))
	} else {
		responseBody["ok"] = false
	}

	return events.APIGatewayProxyResponse{
		Body:       string(prettyResponseBody),
		StatusCode: 200,
	}, nil
}

const (
	base_url = "https://api.telegram.org/bot5566031010:AAEcZzK92iFAYyoCqSck_TEtpGou4AlQEoY/"
)

func removeKeyboard(chatID, messageID string) {

	data := url.Values{}
	data.Set("chat_id", fmt.Sprintf("%v", chatID))
	data.Set("message_id", fmt.Sprintf("%v", messageID))

	replyMarkup := map[string]bool{"remove_keyboard": true}
	replyMarkupJSON, _ := json.Marshal(replyMarkup)

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	client := http.Client{}
	req, _ := http.NewRequest("POST", base_url+"editMessageReplyMarkup", bytes.NewBuffer(replyMarkupJSON))
	req.Header = headers
	req.URL.RawQuery = data.Encode()

	res, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		fmt.Println("Error disabling inline keyboard:", res.Status)
		return
	}

}

func updateTable(url string) (bool, error) {

	tableName := "nyx_feed"
	sess, _ := session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1")},
	)
	svc := dynamodb.New(sess)

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":r": {
				BOOL: aws.Bool(true),
			},
		},
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"link": {
				S: aws.String(strings.Trim(url, `"`)),
			},
		},
		ReturnValues:     aws.String("UPDATED_NEW"),
		UpdateExpression: aws.String("set track = :r"),
	}

	_, err := svc.UpdateItem(input)
	if err != nil {
		log.Fatalf("Got error calling UpdateItem: %s", err)
		return false, err
	}

	return true, nil
}

func main() {
	lambda.Start(handler)
}
