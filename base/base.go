package base

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

// User structure holds authorized users details
type User struct {
	Nickname         string
	Password         string
	TokenType        string
	TokenAuth        string
	TokenSecret      string
	CreationDate     string
	Lastlogin        string
	Email            string
	Active           int
	ValidationString string
	Ports            string
	Server           string
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/")
var simpleLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
var randInit = 0

// MaxAge defines cookie expiration
var MaxAge = 3600 * 24

func randAlphaSlashPlus(n int) string {
	if randInit == 0 {
		rand.Seed(time.Now().UnixNano())
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func randAlpha(n int) string {
	if randInit == 0 {
		rand.Seed(time.Now().UnixNano())
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = simpleLetters[rand.Intn(len(simpleLetters))]
	}
	return string(b)
}

// GenerateAccountACKLink generates account verification link
func GenerateAccountACKLink(length int) string {
	return randAlpha(length)
}

// GenerateAuthToken creates auth token for created user
func GenerateAuthToken(TokenType string, length int) string {
	return randAlphaSlashPlus(length)
}

// HashPassword gets hash from password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash checks given password
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Send some email

var smtpServer string
var smtpAccount string
var smtpPassword string
var bCC string

func initSmtpconfig() error {
	viper.SetConfigName("gatewayconf")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/usr/local/production/config/")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	smtpServer = viper.GetString("SMTP_SERVER") // example: smtp.google.com:587
	smtpAccount = viper.GetString("SMTP_ACCOUNT")
	smtpPassword = viper.GetString("SMTP_PASSWORD")
	bCC = viper.GetString("BCC_ADDRESS")

	return nil
}

// SendEmail provides email function for varied interactions
func SendEmail(email string, subject string, validationString string) {
	var auth smtp.Auth
	err := initSmtpconfig()
	if err != nil {
		Zlog.Errorf("SMTP Config Error %s", err.Error())
	}
	servername := smtpServer
	host, _, _ := net.SplitHostPort(servername)
	shortServer := strings.Split(servername, ":")
	smtpPort := shortServer[1]
	// If I have a short login (aka the login do not contain the domain name from the SMTP server)
	shortName := strings.Split(smtpAccount, "@")
	var from mail.Address
	if len(shortName) > 1 {
		from = mail.Address{"", smtpAccount}
	} else {
		from = mail.Address{"", smtpAccount + "@" + host}
	}
	to := mail.Address{"", email}
	subj := subject
	body := validationString

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body

	// Connect to the SMTP Server
	if len(smtpPassword) > 0 {
		auth = smtp.PlainAuth("", smtpAccount, smtpPassword, host)
	}

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// uncomment the following line to use a pure SSL connection without STARTTLS

	//conn, err := tls.Dial("tcp", servername, tlsconfig)
	conn, err := smtp.Dial(servername)
	if err != nil {
		Zlog.Panicf("SMTP server connection Error %s", err.Error())
	}

	// comment that line to use SSL connection
	if smtpPort != "25" {
		conn.StartTLS(tlsconfig)
	}

	// Auth
	if len(smtpPassword) > 1 {
		if err = conn.Auth(auth); err != nil {
			Zlog.Panicf("Authentication Error %s", err.Error())
		}
	}

	// To && From
	if err = conn.Mail(from.Address); err != nil {
		Zlog.Panicf("SMTP Server MAIL command Error %s", err.Error())
	}

	if err = conn.Rcpt(to.Address); err != nil {
		Zlog.Panicf("SMTP Server RCPT command Error %s", err.Error())
	}

	// Data
	w, err := conn.Data()
	if err != nil {
		Zlog.Panicf("SMTP Server DATA command Error %s", err.Error())
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		Zlog.Panicf("Data writer Error %s", err.Error())
	}

	err = w.Close()
	if err != nil {
		Zlog.Panicf("Writer close Error %s", err.Error())
	}

	conn.Quit()

	if bCC != "" {
		to = mail.Address{"", bCC}
		headers["To"] = to.String()
		// Setup message
		message := ""
		for k, v := range headers {
			message += fmt.Sprintf("%s: %s\r\n", k, v)
		}
		message += "\r\n" + "The following email address has request an account on OSFCI: " + email

		conn, err := smtp.Dial(servername)
		if err != nil {
			Zlog.Panicf("SMTP server connection Error %s", err.Error())
		}

		// comment that line to use SSL connection
		if smtpPort != "25" {
			conn.StartTLS(tlsconfig)
		}

		// Auth
		if len(smtpPassword) > 1 {
			if err = conn.Auth(auth); err != nil {
				Zlog.Panicf("Authentication Error %s", err.Error())
			}
		}

		// To && From
		if err = conn.Mail(from.Address); err != nil {
			Zlog.Panicf("SMTP Server MAIL command Error %s", err.Error())
		}

		if err = conn.Rcpt(to.Address); err != nil {
			Zlog.Panicf("SMTP Server RCPT command Error %s", err.Error())
		}

		// Data
		w, err := conn.Data()
		if err != nil {
			Zlog.Panicf("SMTP Server DATA command Error %s", err.Error())
		}

		_, err = w.Write([]byte(message))
		if err != nil {
			Zlog.Panicf("Data writer Error %s", err.Error())
		}

		err = w.Close()
		if err != nil {
			Zlog.Panicf("Writer close Error %s", err.Error())
		}

		conn.Quit()

	}

	if err != nil {
		Zlog.Errorf("SMTP Error %s", err.Error())
	}

}

// Request handler
func Request(method string, resURI string, Path string, Data string, content []byte, query string, Key string, SecretKey string) (*http.Response, error) {

	client := &http.Client{}

	myDate := time.Now().UTC().Format(http.TimeFormat)
	myDate = strings.Replace(myDate, "GMT", "+0000", -1)
	var req *http.Request
	if content != nil {
		req, _ = http.NewRequest(method, resURI, bytes.NewReader(content))
	} else {
		req, _ = http.NewRequest(method, resURI, nil)
	}

	stringToSign := method + "\n\n" + Data + "\n" + myDate + "\n" + Path

	mac := hmac.New(sha1.New, []byte(SecretKey))
	mac.Write([]byte(stringToSign))
	expectedMAC := mac.Sum(nil)
	signature := base64.StdEncoding.EncodeToString(expectedMAC)

	req.Header.Set("Authorization", "AWS "+Key+":"+signature)
	req.Header.Set("Date", myDate)
	req.Header.Set("Content-Type", Data)
	if len(content) > 0 {
		req.ContentLength = int64(len(content))
	}

	req.URL.RawQuery = query

	// That is a new request so let's do it
	var response *http.Response
	var err error
	response, err = client.Do(req)
	return response, err

}

// HTTPGetRequest handles some HTTP request
// Get request to the storage backend
func HTTPGetRequest(request string) string {
	resp, err := http.Get(request)
	if err != nil {
		Zlog.Fatalf("HTTP GET Error %s", err.Error())
		return ""
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		Zlog.Fatalf("HTTP GET response read Error %s", err.Error())
	}
	return (string(body))
}

// HTTPDeleteRequest handles Delete request to backend
func HTTPDeleteRequest(request string) {
	client := &http.Client{}
	content := []byte{0}
	httprequest, err := http.NewRequest("DELETE", request, bytes.NewReader(content))
	httprequest.ContentLength = 0
	response, err := client.Do(httprequest)
	if err != nil {
		Zlog.Fatalf("HTTP DELETE Error %s", err.Error())
	} else {
		defer response.Body.Close()
		_, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Zlog.Fatalf("HTTP DELETE response read Error %s", err.Error())
		}
	}
}

// HTTPPutRequest handles Put request to the storage backend
func HTTPPutRequest(request string, content []byte, contentType string) string {
	client := &http.Client{}
	httprequest, err := http.NewRequest("PUT", request, bytes.NewReader(content))
	httprequest.Header.Set("Content-Type", contentType)
	httprequest.ContentLength = int64(len(content))
	response, err := client.Do(httprequest)
	if err != nil {
		Zlog.Fatalf("HTTP PUT Error %s", err.Error())
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			Zlog.Fatalf("HTTP PUT response read Error %s", err.Error())
		}
		return string(contents)
	}
	return ""
}

// HTTPGetBody handles request body for redirects
func HTTPGetBody(r *http.Request) []byte {
	buf, _ := ioutil.ReadAll(r.Body)
	rdr1 := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))
	b := new(bytes.Buffer)
	b.ReadFrom(rdr1)
	r.Body = rdr2
	return (b.Bytes())
}

// CheckURLExists handles checks ia URL exists
func CheckURLExists(request string) bool {
	_, err := http.Get(request)
	if err != nil {
		Zlog.Warnf("HTTP GET Error %s", err.Error())
		return false
	}
	return true
}
