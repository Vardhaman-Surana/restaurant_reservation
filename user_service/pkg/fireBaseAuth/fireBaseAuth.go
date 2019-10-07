package fireBaseAuth

import (
	"bytes"
	"context"
	"encoding/json"
	firebase "firebase.google.com/go"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)
const verifyCustomTokenURL = "https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=AIzaSyAfwNKFlei2Ys6yGslxrunrV64mpkzzwpE"


var app *firebase.App

func InitFireBase()error{
	var err error
	app, err = firebase.NewApp(context.Background(), nil)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}
	if err!=nil{
		log.Fatalf("Can not initialize firebase:%v",err)
		return err
	}
	return nil
}

func CreateToken(ctx context.Context,id string)(token string,err error){
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}
	token, err = client.CustomToken(ctx,id)
	if err != nil {
		log.Fatalf("error minting custom token: %v", err)
	}

	log.Printf("Got custom token: %v", token)
	return token,err
}

func VerifyToken(ctx context.Context,token string)(string,error){
	// [START verify_id_token_golang]
	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	tkn, err := client.VerifyIDToken(ctx,token)
	if err != nil {
		log.Fatalf("error verifying ID token: %v", err)
	}
	log.Printf("Verified ID token: %v", tkn)
	return tkn.UID,err
}


func SignInWithCustomToken(token string) (string, error) {
	req, err := json.Marshal(map[string]interface{}{
		"token":             token,
		"returnSecureToken": true,
	})
	if err != nil {
		return "", err
	}

	resp, err := postRequest(verifyCustomTokenURL, req)
	if err != nil {
		return "", err
	}
	var respBody struct {
		IDToken string `json:"idToken"`
	}
	if err := json.Unmarshal(resp, &respBody); err != nil {
		return "", err
	}
	return respBody.IDToken, err
}

func postRequest(url string, req []byte) ([]byte, error) {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _:=ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected http status: %v", string(b))
	}
	return ioutil.ReadAll(resp.Body)
}
