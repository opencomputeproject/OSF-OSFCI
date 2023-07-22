package main

import (
	"base/base"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	verifier "github.com/okta/okta-jwt-verifier-golang"
	"github.com/spf13/viper"
)

var (
	state = generateState()
	nonce = ""
)

// StorageURI is read from config
var StorageURI string

// StorageTCPPORT is read from config
var StorageTCPPORT string

// CredentialURI is read from config
var CredentialURI string

// ClientID for Okta
var ClientID string

// Issuer URL for Okta
var Issuer string

// ClientSecret for Okta
var ClientSecret string

// RedirectURL Signin Redirect URL defined in the Okta
var RedirectURL string

// LogoutRedirectURL Logout Redirect URL defined in the Okta
var LogoutRedirectURL string

type cacheEntry struct {
	Nickname string
	Cookie   string
	Expire   time.Time
}

var cache []cacheEntry

// Struct to store the new user details
type authUser struct {
	Nickname    string
	TokenType   string
	TokenAuth   string
	TokenSecret string
	Email       string
	TokenID     string
	AccessToken string
}

// UserDB To store the hash of user details after login
var UserDB map[string]*authUser

// Upercase is mandatory for JSON library parsing

type userPublic struct {
	Nickname         string
	NicknameRW       string
	NicknameLABEL    string
	TokenType        string
	TokenTypeRW      string
	TokenAuth        string
	TokenAuthRW      string
	TokenSecret      string
	TokenSecretLABEL string
	TokenSecretRW    string
	CreationDate     string
	CreationDateRW   string
	Lastlogin        string
	LastloginRW      string
	Email            string
	EmailRW          string
	EmailLABEL       string
}

// Exchange - Token structure returned by Okta
type Exchange struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	AccessToken      string `json:"access_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	Scope            string `json:"scope,omitempty"`
	IdToken          string `json:"id_token,omitempty"`
}

// Initialize User config
func initUserconfig() error {
	viper.SetConfigName("gatewayconf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/usr/local/production/config/")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	//StorageURI set from config file
	StorageURI = viper.GetString("STORAGE_URI")

	//StorageTCPPORT set from config file
	StorageTCPPORT = viper.GetString("STORAGE_TCPPORT")
	CredentialURI = viper.GetString("CREDENTIALS_TCPPORT")
	return nil
}

func initAuthconfig() error {
	UserDB = make(map[string]*authUser)
	config := viper.New()
	config.SetConfigName("hpeauth")
	config.SetConfigType("yaml")
	config.AddConfigPath(os.Getenv("CONFIG_PATH"))

	err := config.ReadInConfig()
	if err != nil {
		return err
	}
	//StorageURI set from config file

	ClientID = config.GetString("CLIENT_ID")
	ClientSecret = config.GetString("CLIENT_SECRET")
	Issuer = config.GetString("ISSUER")
	RedirectURL = config.GetString("REDIRECT_URL")
	LogoutRedirectURL = config.GetString("SIGNOUT_REDIRECT_URL")
	return nil
}

func verifyToken(t string) (*verifier.Jwt, error) {
	tv := map[string]string{}
	tv["nonce"] = nonce
	tv["aud"] = ClientID
	jv := verifier.JwtVerifier{
		Issuer:           Issuer,
		ClaimsToValidate: tv,
	}

	result, err := jv.New().VerifyIdToken(t)
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	if result != nil {
		return result, nil
	}

	return nil, fmt.Errorf("token could not be verified: %s", "")
}

func generateState() string {
	// Generate a random byte array for state parameter
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func generateNonce() (string, error) {
	nonceBytes := make([]byte, 32)
	_, err := rand.Read(nonceBytes)
	if err != nil {
		return "", fmt.Errorf("could not generate nonce")
	}

	return base64.URLEncoding.EncodeToString(nonceBytes), nil
}

func userExist(username string) bool {
	// We must call the storage backend with the username
	var result string
	// that must be an http request instead of a vejmarie
	result = base.HTTPGetRequest("http://" + StorageURI + StorageTCPPORT + "/user/" + username)
	if result == "Error" {
		fmt.Printf("User doesn't exist\n")
		return false
	}
	return true
}

func userGetInfo(nickname string) *userPublic {
	// We must call the storage backend service to get access to the resource
	// We could have a bucket / fileid approach which could be translated into flat file
	// or database management
	var tempValue *base.User
	var returnValue *userPublic
	var result string
	if userExist(nickname) {
		result = base.HTTPGetRequest("http://" + StorageURI + StorageTCPPORT + "/user/" + nickname)
		tempValue = new(base.User)
		json.Unmarshal([]byte(result), tempValue)
		returnValue = new(userPublic)
		returnValue.Nickname = tempValue.Nickname
		returnValue.NicknameRW = "0"
		returnValue.NicknameLABEL = "This is your unique identifier. It will appeared within your publications and used to refer you as author. It is visible to any other users."
		returnValue.TokenType = tempValue.TokenType
		returnValue.TokenTypeRW = "0"
		returnValue.TokenAuth = tempValue.TokenAuth
		returnValue.TokenAuthRW = "0"
		returnValue.TokenSecret = tempValue.TokenSecret
		returnValue.TokenSecretLABEL = "TokenType, TokenAuth and TokenSecret are private values that you shouldn't share with anybody. They are automatically assigned to you as to provide you unique authentication capabilities to this service."
		returnValue.TokenSecretRW = "0"
		returnValue.CreationDate = tempValue.CreationDate
		returnValue.CreationDateRW = "0"
		returnValue.Lastlogin = tempValue.Lastlogin
		returnValue.LastloginRW = "0"
		returnValue.Email = tempValue.Email
		returnValue.EmailLABEL = "Your primary email address. It won't be shared with anybody. Warning your email address must be verified each time you change it. During that process your account is disabled and can't be recovered without contacting us."
		returnValue.EmailRW = "1"
	}

	return returnValue
}

func userGetInternalInfo(nickname string) *base.User {
	// We must call the storage backend service to get access to the resource
	// We could have a bucket / fileid approach which could be translated into flat file
	// or database management
	var returnValue *base.User
	var result string
	if userExist(nickname) {
		base.Zlog.Infof("User exists")
		result = base.HTTPGetRequest("http://" + StorageURI + StorageTCPPORT + "/user/" + nickname)
		returnValue = new(base.User)
		json.Unmarshal([]byte(result), returnValue)
	}
	return returnValue
}

func updateAccount(username string, w http.ResponseWriter, r *http.Request) bool {
	var updatedData *base.User
	var serverReturn string
	serverReturn = ""
	type accountUpdate struct {
		Email           string
		CurrentPassword string
		NewPassword0    string
		NewPassword1    string
	}
	exist := userExist(username)
	if !exist {
		fmt.Fprint(w, "Error")
		return false
	}
	updatedData = userGetInternalInfo(username)
	var getJSON = base.HTTPGetBody(r)
	var newData accountUpdate

	// We have to unMarshal the body to update the data

	_ = json.Unmarshal(getJSON, &newData)

	// So now let's run some comparaison
	if updatedData.Active == 0 {
		http.Error(w, "401 User not activated Please check email", 401)
		return false
	}

	if newData.CurrentPassword != "undefined" {
		if !base.CheckPasswordHash(newData.CurrentPassword, updatedData.Password) {
			w.Write([]byte("error password"))
			return false
		}
		// we are good to update the password and log off the user
		// but only if the size is bigger than 0 !
		if newData.NewPassword0 != "undefined" {
			updatedData.Password, _ = base.HashPassword(newData.NewPassword0)
			b, _ := json.Marshal(updatedData)
			base.HTTPPutRequest("http://"+StorageURI+StorageTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
			serverReturn = serverReturn + "password"
		}
	}

	// If the email address are different
	if updatedData.Email != newData.Email {
		// We must put the account into an inactive mode as long as the new email has not been validated
		// We must renew the email check account
		updatedData.Email = newData.Email
		updatedData.Active = 0
		// we change the Validation string and send the email
		updatedData.ValidationString = base.GenerateAccountACKLink(24)
		b, _ := json.Marshal(updatedData)
		base.HTTPPutRequest("http://"+StorageURI+StorageTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
		base.SendEmail(updatedData.Email, "Account activation - Action required",
			"Please click the following link as to validate your account https://"+
				r.Host+"/user/"+updatedData.Nickname+"/validate_user/"+updatedData.ValidationString)
		updatedData = nil
		serverReturn = serverReturn + "email"
	}

	// If the Password is modified we must validate that the previous password has been properly typed in
	w.Write([]byte(serverReturn))
	return true

}

func createUser(username string, w http.ResponseWriter, r *http.Request) bool {
	var updatedData *base.User
	exist := userExist(username)
	if exist {
		fmt.Fprint(w, "Error")
		return false
	}

	updatedData = new(base.User)
	updatedData.Nickname = username
	updatedData.Email = r.FormValue("email")
	if base.ValidateDomain(updatedData.Email) == false {
		fmt.Fprint(w, "Error")
		return false
	}

	// this is a creation
	updatedData.TokenAuth = base.GenerateAccountACKLink(20)
	updatedData.TokenSecret = base.GenerateAuthToken("mac", 40)
	updatedData.TokenType = "mac"
	updatedData.CreationDate = string(time.Now().Format(time.RFC1123Z))
	updatedData.Password, _ = base.HashPassword(r.FormValue("password"))
	updatedData.Lastlogin = ""
	updatedData.Active = 0
	updatedData.ValidationString = base.GenerateAccountACKLink(24)
	b, _ := json.Marshal(updatedData)
	base.HTTPPutRequest("http://"+StorageURI+StorageTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
	base.SendEmail(updatedData.Email, "Account activation - Action required",
		"Please click the following link as to validate your account https://"+
			r.Host+"/user/"+updatedData.Nickname+"/validate_user/"+updatedData.ValidationString)
	updatedData = nil
	return true

}

func updateAvatar(username string, w http.ResponseWriter, r *http.Request) bool {
	// We must store the body content within the avatar file of the end user
	exist := userExist(username)
	if !exist {
		fmt.Fprint(w, "Error")
		return false
	}
	base.HTTPPutRequest("http://"+StorageURI+StorageTCPPORT+"/user/"+username, base.HTTPGetBody(r), "image/jpg")
	return true
}

func getAvatar(username string, w *http.ResponseWriter) {
	exist := userExist(username)
	if !exist {
		fmt.Fprint(*w, "Error")
		return
	}
	if base.CheckURLExists("http://" + StorageURI + StorageTCPPORT + "/user/" + username + "/avatar") {
		(*w).Write([]byte(base.HTTPGetRequest("http://" + StorageURI + StorageTCPPORT + "/user/" + username + "/avatar")))
	}
}

func sendPasswordResetLink(username string, w http.ResponseWriter, r *http.Request) bool {
	var updatedData *base.User
	exist := userExist(username)
	if !exist {
		fmt.Fprint(w, "Error")
		return false
	}
	updatedData = userGetInternalInfo(username)
	updatedData.ValidationString = base.GenerateAccountACKLink(24)
	// The user can't be active as long as we do not have reset the password
	updatedData.Active = 0
	b, _ := json.Marshal(updatedData)
	base.HTTPPutRequest("http://"+StorageURI+StorageTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
	base.SendEmail(updatedData.Email, "Account password reset - Action required",
		"Please click the following link as to update  your password https://"+
			r.Host+"/user/"+updatedData.Nickname+"/reset_password/"+updatedData.ValidationString)
	updatedData = nil
	return true

}

func resetPassword(username string, w http.ResponseWriter, r *http.Request) bool {
	var updatedData *base.User
	exist := userExist(username)
	if !exist {
		fmt.Fprint(w, "Error")
		return false
	}
	updatedData = userGetInternalInfo(username)
	if updatedData.ValidationString != r.FormValue("validation") {
		fmt.Fprint(w, "Error")
		return false
	}
	updatedData.ValidationString = ""
	updatedData.Password, _ = base.HashPassword(r.FormValue("password"))
	updatedData.Active = 1
	b, _ := json.Marshal(updatedData)
	base.HTTPPutRequest("http://"+StorageURI+StorageTCPPORT+"/user/"+updatedData.Nickname, b, "application/json")
	return true
}

func validateUser(username string, validationstring string) bool {
	var updatedData *base.User
	// We  must check if the user exist
	exist := userExist(username)
	if !exist {
		return false
	}
	// We must read the user data and update the content of it
	updatedData = userGetInternalInfo(username)
	// We must check that the validation string is a match
	if updatedData.ValidationString != validationstring {
		return false
	}
	updatedData.Active = 1

	// We write back the data
	c, _ := json.Marshal(updatedData)
	base.HTTPPutRequest("http://"+StorageURI+StorageTCPPORT+"/user/"+updatedData.Nickname, c, "application/json")

	// And return positively
	return true
}

func deleteUser(username string, w http.ResponseWriter, r *http.Request) bool {
	// We delete the user by a direct call to the storage subsystem
	var updatedData *base.User
	// I am receiving the password within the http body of the delete request
	type accountDelete struct {
		CurrentPassword string
		DeleteData      string
	}
	var newData accountDelete
	var getJSON = base.HTTPGetBody(r)
	_ = json.Unmarshal(getJSON, &newData)
	base.Zlog.Infof("Deleting the user: %s", username)
	if newData.DeleteData == "true" {
	} else {
	}
	updatedData = userGetInternalInfo(username)
	// if the received password is not the one of the end user we can't erase it's account
	// might be a browser hack
	if !base.CheckPasswordHash(newData.CurrentPassword, updatedData.Password) {
		w.Write([]byte("error password"))
		return false
	}

	base.Zlog.Infof("Confirm Deleting the user: %s", updatedData.Nickname)
	// Just need to disable the account by unactivating it
	// It could be recovered by resetting the password
	updatedData.Active = 0
	//c, _ := json.Marshal(updatedData)
	//base.HTTPPutRequest("http://"+StorageURI+StorageTCPPORT+"/user/"+updatedData.Nickname, c, "application/json")
	base.HTTPDeleteRequest("http://" + StorageURI + StorageTCPPORT + "/user/" + updatedData.Nickname)
	if newData.DeleteData == "true" {
		base.HTTPDeleteRequest("http://" + StorageURI + StorageTCPPORT + "/user/" + updatedData.Nickname + "/delete_user_data")
	}
	// And return positively
	return true
}

func getSessionID(username string) string {
	// We need to save the cookie into the user database (TODO)
	// Is the user into the cache
	for _, entry := range cache {
		if entry.Nickname == username {
			if entry.Expire.After(time.Now()) {
				// Ok the Cookie is not expired
				// We can return it and extend the lifecycle
				entry.Expire = time.Now().Add(time.Second * time.Duration(base.MaxAge))
				return (entry.Cookie)
			}
		}
	}

	// ok we must add an entry

	var newEntry cacheEntry
	newEntry.Nickname = username
	newEntry.Expire = time.Now().Add(time.Second * time.Duration(base.MaxAge))
	Data := make([]byte, 32)
	io.ReadFull(rand.Reader, Data)
	cookie := base64.URLEncoding.EncodeToString(Data)
	newEntry.Cookie = cookie
	cache = append(cache, newEntry)
	return (newEntry.Cookie)

}

func getOpenBMC(username string, w http.ResponseWriter, recipe string) {
	client := &http.Client{}
	var req *http.Request
	req, _ = http.NewRequest("GET", "http://"+StorageURI+StorageTCPPORT+"/user/"+username+"/get_bmc_firmware/"+recipe, nil)
	response, _ := client.Do(req)
	buf, _ := ioutil.ReadAll(response.Body)
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.Write(buf)
}

func getOpenBMCBuildLog(username string, w http.ResponseWriter, recipe string) {
	client := &http.Client{}
	var req *http.Request
	req, _ = http.NewRequest("GET", "http://"+StorageURI+StorageTCPPORT+"/user/"+username+"/get_bmc_firmware_build_log/"+recipe, nil)
	response, _ := client.Do(req)
	buf, _ := ioutil.ReadAll(response.Body)
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.Write(buf)
}

func getLinuxBoot(username string, w http.ResponseWriter, recipe string) {
	client := &http.Client{}
	var req *http.Request
	req, _ = http.NewRequest("GET", "http://"+StorageURI+StorageTCPPORT+"/user/"+username+"/get_firmware/"+recipe, nil)
	response, _ := client.Do(req)
	buf, _ := ioutil.ReadAll(response.Body)
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.Write(buf)
}

func getLinuxBootBuildLog(username string, w http.ResponseWriter, recipe string) {
	client := &http.Client{}
	var req *http.Request
	req, _ = http.NewRequest("GET", "http://"+StorageURI+StorageTCPPORT+"/user/"+username+"/get_firmware_build_log/"+recipe, nil)
	response, _ := client.Do(req)
	buf, _ := ioutil.ReadAll(response.Body)
	w.Header().Set("Content-Length", strconv.Itoa(len(buf)))
	w.Write(buf)
}

func authHpe(username string, password string, w http.ResponseWriter) {
	payload := map[string]string{"username": username, "password": password}
	byts, _ := json.Marshal(payload)
	url := "https://auth.hpe.com/api/v1/authn"
	req, err := http.NewRequest("POST", url, bytes.NewReader(byts))
	if err != nil {
		base.Zlog.Errorf(err.Error())
		http.Error(w, "Authentication failed", 401)
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		base.Zlog.Errorf(err.Error())
		http.Error(w, "Authentication failed", 401)
		return
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		base.Zlog.Errorf(err.Error())
		http.Error(w, "Authentication failed", 401)
		return
	}
	base.Zlog.Infof(string(body))
	response := make(map[string]interface{})
	json.Unmarshal(body, &response)
	_, is_logged := response["status"]
	if is_logged && response["status"].(string) == "SUCCESS" {
		base.Zlog.Infof(response["status"].(string))
		profile := response["_embedded"].(map[string]interface{})["user"].(map[string]interface{})["profile"].(map[string]interface{})
		var user = new(authUser)
		user.Nickname = profile["firstName"].(string) + profile["lastName"].(string)
		user.Email = profile["login"].(string)
		user.TokenAuth = base.GenerateAccountACKLink(20)
		user.TokenSecret = base.GenerateAuthToken("mac", 40)
		user.TokenType = "mac"
		user.AccessToken = response["stateToken"].(string)
		user.TokenID = response["sessionToken"].(string)
		UserDB[user.Email] = user
		returnData := map[string]string{}
		returnData["accessKey"] = user.TokenAuth
		returnData["secretKey"] = user.TokenSecret
		returnValue, _ := json.Marshal(returnData)
		sessionid := getSessionID(user.Email)
		cookie := http.Cookie{Name: "osfci_cookie", Value: sessionid, Path: "/", HttpOnly: true, MaxAge: int(base.MaxAge)}
		http.SetCookie(w, &cookie)
		fmt.Fprintf(w, string(returnValue))
	} else {
		base.Zlog.Errorf("Authentication failed")
		http.Error(w, "Authentication failed", 401)
	}
	defer resp.Body.Close()
}

func userCallback(w http.ResponseWriter, r *http.Request) {
	var username, command string
	var recipe string

	path := strings.Split(r.URL.Path, "/")
	if len(path) < 3 {
		http.Error(w, "401 Malformed URI", 401)
		return
	}
	username = path[2]
	if len(path) >= 4 {
		command = path[3]
	}
	switch r.Method {
	case http.MethodGet:
		switch command {
		case "validate_user":
			// got a validation link ....
			// we have to accept user activation
			// First check if the account exist
			// if yes we must get the data, compare the link and if a match
			// activate the user allowing a call to the API to get the connection token
			if !validateUser(username, path[4]) {
				http.Error(w, "401 Validation string error", 401)
			} else {
				// We just need to display the login page
				// One of the issue is that it is a dynamic page
				// We need to do it through the app.js
				// and load the script in a way it can detect the redirection
				http.Redirect(
					w, r,
					"https://"+r.Host+"/ci/?loginValidated=1",
					http.StatusMovedPermanently,
				)
			}
		case "reset_password":
			// We have to validate the user, then display the right return page
			if !validateUser(username, path[4]) {
				http.Error(w, "401 Validation string error", 401)
			} else {
				http.Redirect(
					w, r,
					"https://"+r.Host+"/ci/?reset_password=1&username="+username+"&validation="+path[4],
					http.StatusMovedPermanently,
				)
			}
		case "userGetInternalInfo":
			var result *base.User
			// Serve the resource.
			fmt.Printf("Requesting %s\n", username)
			if userExist(username) {
				result = userGetInternalInfo(username)
				b, _ := json.Marshal(*result)
				fmt.Fprint(w, string(b))
				return
			}
			user, ok := UserDB[username]
			if ok == true {
				result = new(base.User)
				result.Nickname = username
				result.Email = username
				result.TokenAuth = user.TokenAuth
				result.TokenSecret = user.TokenSecret
				result.TokenType = user.TokenType
				result.CreationDate = ""
				result.Password = ""
				result.Lastlogin = ""
				result.Active = 1
				result.ValidationString = ""
				b, _ := json.Marshal(*result)
				fmt.Fprint(w, string(b))
			}
			return
		case "userGetInfo":
			var result *userPublic
			// Serve the resource.
			result = userGetInfo(username)
			b, _ := json.Marshal(*result)
			fmt.Fprint(w, string(b))

		case "getAvatar":
			getAvatar(username, &w)
		case "getOpenBMC":
			recipe = path[4]
			getOpenBMC(username, w, recipe)
			base.Zlog.Infof("%s Downloaded OpenBMC image via IP: %s", username, base.GetClientIP(r))
		case "getLinuxBoot":
			recipe = path[4]
			getLinuxBoot(username, w, recipe)
			base.Zlog.Infof("%s Downloaded LinuxBoot image via IP: %s", username, base.GetClientIP(r))
		case "getOpenBMCLog":
			recipe = path[4]
			getOpenBMCBuildLog(username, w, recipe)
			base.Zlog.Infof("%s Downloaded OpenBMC log via IP: %s", username, base.GetClientIP(r))
		case "getLinuxBootLog":
			recipe = path[4]
			getLinuxBootBuildLog(username, w, recipe)
			base.Zlog.Infof("%s Downloaded LinuxBoot log via IP: %s", username, base.GetClientIP(r))
		case "authverify":
			code := r.URL.Query().Get("code")
			if code == "" {
				base.Zlog.Infof("The code was not returned or is not accessible")
				return
			}
			authHeader := base64.StdEncoding.EncodeToString([]byte(ClientID + ":" + ClientSecret))
			q := r.URL.Query()
			q.Add("grant_type", "authorization_code")
			q.Set("code", code)
			q.Add("redirect_uri", RedirectURL)

			url := Issuer + "/v1/token?" + q.Encode()

			req, _ := http.NewRequest("POST", url, bytes.NewReader([]byte("")))
			h := req.Header
			h.Add("Authorization", "Basic "+authHeader)
			h.Add("Accept", "application/json")
			h.Add("Content-Type", "application/x-www-form-urlencoded")
			h.Add("Connection", "close")
			h.Add("Content-Length", "0")

			client := &http.Client{}
			resp, _ := client.Do(req)
			body, _ := ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			var exchange Exchange
			json.Unmarshal(body, &exchange)
			if exchange.Error != "" {
				base.Zlog.Infof(exchange.Error)
				base.Zlog.Infof(exchange.ErrorDescription)
				return
			}
			_, verificationError := verifyToken(exchange.IdToken)
			if verificationError != nil {
				base.Zlog.Warnf(verificationError.Error())
				return
			}
			reqURL := Issuer + "/v1/userinfo"

			req, _ = http.NewRequest("GET", reqURL, bytes.NewReader([]byte("")))
			h = req.Header
			h.Add("Authorization", "Bearer "+exchange.AccessToken)
			h.Add("Accept", "application/json")

			client = &http.Client{}
			resp, _ = client.Do(req)
			body, _ = ioutil.ReadAll(resp.Body)
			defer resp.Body.Close()
			profile := make(map[string]string)
			json.Unmarshal(body, &profile)
			base.Zlog.Infof(profile["email"])

			var user = new(authUser)
			user.Nickname = profile["name"]
			user.Email = profile["email"]
			user.TokenAuth = base.GenerateAccountACKLink(20)
			user.TokenSecret = base.GenerateAuthToken("mac", 40)
			user.TokenType = "mac"
			user.AccessToken = exchange.AccessToken
			user.TokenID = exchange.IdToken
			UserDB[user.Email] = user
			cookie := http.Cookie{Name: "OSFCIAUTH", Value: user.Email, Path: "/", HttpOnly: true, MaxAge: int(30)}
			http.SetCookie(w, &cookie)
			http.Redirect(
				w, r,
				"https://"+r.Host+"/ci/?is_authenicated=1",
				http.StatusFound,
			)
			base.Zlog.Infof("End of User Redirect")
		case "authtoken":
			base.Zlog.Infof("Inside Auth token")
			returnData := make(map[string]interface{})
			usercookie, err := r.Cookie("OSFCIAUTH")
			if err != nil || usercookie.Value == "" {
				returnData["Error"] = "Unable to fetch User profile"
				returnValue, _ := json.Marshal(returnData)
				fmt.Fprintf(w, string(returnValue))
				return
			}
			base.Zlog.Infof(usercookie.Value)
			user, ok := UserDB[usercookie.Value]
			if ok == false {
				returnData["Error"] = "Error:Unable to fetch User profile"
				returnValue, _ := json.Marshal(returnData)
				fmt.Fprintf(w, string(returnValue))
				return
			}

			returnData["accessKey"] = user.TokenAuth
			returnData["secretKey"] = user.TokenSecret
			returnData["username"] = usercookie.Value
			returnData["osfciauth"] = true
			returnValue, _ := json.Marshal(returnData)
			sessionid := getSessionID(usercookie.Value)
			cookie := http.Cookie{Name: "osfci_cookie", Value: sessionid, Path: "/", HttpOnly: true, MaxAge: int(base.MaxAge)}
			http.SetCookie(w, &cookie)
			fmt.Fprintf(w, string(returnValue))
		case "verify_user":
			// We have to validate the user, then display the right return page
			returndata := make(map[string]interface{})
			exist := userExist(username)
			returndata["Exists"] = 0
			if exist {
				returndata["Exists"] = 1
				returnValue, _ := json.Marshal(returndata)
				w.Write([]byte(returnValue))
				return
			}
			nonce, _ = generateNonce()
			var redirectPath string

			q := r.URL.Query()
			q.Add("client_id", ClientID)
			q.Add("response_type", "code")
			q.Add("response_mode", "query")
			q.Add("scope", "openid profile email")
			q.Add("login_hint", username)
			q.Add("redirect_uri", RedirectURL)
			q.Add("state", state)
			q.Add("prompt", "login")
			q.Add("nonce", nonce)

			redirectPath = Issuer + "/v1/authorize?" + q.Encode()
			base.Zlog.Infof(redirectPath)
			returndata["Redirect"] = redirectPath
			returnValue, _ := json.Marshal(returndata)
			w.Write([]byte(returnValue))
			return
		case "authlogout":
			returndata := make(map[string]interface{})
			base.Zlog.Infof(username)
			user, ok := UserDB[username]
			if ok == false {
				returndata["Error"] = "Unable to find the user"
			} else {
				q := r.URL.Query()
				q.Add("id_token_hint", user.TokenID)
				q.Add("post_logout_redirect_uri", LogoutRedirectURL)
				redirectPath := Issuer + "/v1/logout?" + q.Encode()
				returndata["Redirect"] = redirectPath
			}
			returnValue, _ := json.Marshal(returndata)
			w.Write([]byte(returnValue))
		default:
		}
	case http.MethodPut:
		// Update an existing record.
		switch command {
		case "updateAvatar":
			updateAvatar(username, w, r)
		case "updateAccount":
			updateAccount(username, w, r)
		default:
			http.Error(w, "401 Unknown user command", 401)
			return
		}
	case http.MethodPost:
		// Ok I am getting there the various parameters to log a user
		switch command {
		case "get_token":
			// We must get the user info and validate the password sent
			// if the user doesn't have any API Token
			// we have to generate it !
			// if the user doesn't exist we need to deny the request
			base.Zlog.Infof("GetToken: %s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
			password := r.FormValue("password")
			var result *base.User
			result = userGetInternalInfo(username)
			base.Zlog.Infof("User check done")
			if result == nil {
				base.Zlog.Infof("Trying to authenticate using HPE Auth")
				authHpe(username, password, w)
				return
			}
			if !base.CheckPasswordHash(password, result.Password) {
				http.Error(w, "401 Password error", 401)
				base.Zlog.Infof("Password error: %s - %s %s %s", r.RemoteAddr, r.Proto, r.Method, r.URL.RequestURI())
				return
			}
			if result.Active == 0 {
				http.Error(w, "401 User not activated Please check email", 401)
				return
			}
			// We have the right password !
			// So, we need to send the secret and access token
			// as the end user could login the to the API
			// and load the right page !
			returnValue := " { \"accessKey\" : \"" + result.TokenAuth +
				"\", \"secretKey\" : \"" + result.TokenSecret + "\" }"
			result.Lastlogin = string(time.Now().Format(time.RFC1123Z))
			b, _ := json.Marshal(result)
			base.HTTPPutRequest("http://"+StorageURI+StorageTCPPORT+"/user/"+result.Nickname, b, "application/json")

			// As the user might be willing to use OpenBMC we need to send him also a SESSION ID cookie
			// which will be the only way to track him/her as we eveolve from a single app web base
			// platform to a multiple one (our website and the OpenBMC one)
			sessionid := getSessionID(result.Nickname)
			// We need to send back the cookie to the client
			cookie := http.Cookie{Name: "osfci_cookie", Value: sessionid, Path: "/", HttpOnly: true, MaxAge: int(base.MaxAge)}
			http.SetCookie(w, &cookie)
			fmt.Fprintf(w, string(returnValue))
		case "create_user":
			createUser(username, w, r)
		case "generate_password_lnk_rst":
			sendPasswordResetLink(username, w, r)
		case "reset_password":
			resetPassword(username, w, r)
		default:
			http.Error(w, "401 Unknown user command\n", 401)

		}
	case http.MethodDelete:
		// Remove the record.
		deleteUser(username, w, r)
	default:
		http.Error(w, "401 Unknown request\n", 401)
	}
}

// Default Intialize
func init() {

	config := base.Configuration{
		EnableConsole:     false,                                 //print output on the console, Good for debugging in local
		ConsoleLevel:      base.Debug,                            //Debug level log
		ConsoleJSONFormat: false,                                 //Console log in JSON format, false will print in raw format on console
		EnableFile:        true,                                  // Logging in File
		FileLevel:         base.Info,                             // File log level
		FileJSONFormat:    true,                                  // File JSON Format, False will print in file in raw Format
		FileLocation:      "/usr/local/production/logs/user.log", //File location where log needs to be appended
	}

	err := base.NewLogger(config)
	if err != nil {
		base.Zlog.Fatalf("Could not instantiate user log %s", err.Error())
	}
	base.Zlog.Infof("Starting user logger...")
}

func main() {
	base.Zlog.Infof("Starting user...")
	// http to https redirection

	err := initUserconfig()
	if err != nil {
		base.Zlog.Fatalf("Initialization error: %s", err.Error())
	}

	// Initializing the list of  blocked Domain and IP
	err = base.InitProhibitedIPs()
	if err != nil {
		base.Zlog.Warnf("IP filter initialization error: %s", err.Error())
	}

	// Initializing the list of  blocked Domain and IP
	err = initAuthconfig()
	if err != nil {
		base.Zlog.Warnf("Auth initialization error: %s", err.Error())
	}

	mux := http.NewServeMux()
	// Serve one page site dynamic pages
	mux.HandleFunc("/user/", userCallback)
	if err := http.ListenAndServe(CredentialURI, mux); err != http.ErrServerClosed {
		base.Zlog.Fatalf("User service error: %s", err.Error())
	}
}
