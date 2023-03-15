package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const ()

type PresignedURLResponse struct {
	URL string `json:"url"`
}

func goDotEnvVariable(key string) string {

	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	result := os.Getenv(key)

	if len(result) == 0 {
		log.Fatalf("%s environment variable not set. You can also add to local .env file", key)
	}
	return result
}

var apiKey = goDotEnvVariable("S3_PROXY_API_KEY")
var apiURL = goDotEnvVariable("API_URL")

func main() {
	apiKey := goDotEnvVariable("S3_PROXY_API_KEY")
	if len(apiKey) == 0 {
		log.Fatal("S3_PROXY_API_KEY environment variable not set.  You can add to local .env file")
	}

	if len(os.Args) < 2 {
		displayHelp()
		os.Exit(1)
	}

	operation := os.Args[1]

	os.Args = append(os.Args[:1], os.Args[2:]...) // Remove operation from the arguments

	flagSet := flag.NewFlagSet("s3cli", flag.ExitOnError)

	bucket := flagSet.String("bucket", "", "Bucket name")
	key := flagSet.String("key", "", "Object key")
	filename := flagSet.String("file", "", "File to upload or download")
	presigned := flagSet.Bool("presigned", false, "Use presigned URL for get and put operations")

	flagSet.Parse(os.Args[1:])

	switch operation {
	case "list":
		listBuckets()
	case "get":
		getObject(*bucket, *key, *filename, *presigned)
	case "put":
		putObject(*bucket, *key, *filename, *presigned)
	case "delete":
		deleteObject(*bucket, *key)
	default:
		displayHelp()
		os.Exit(0)
	}
}

func listBuckets() {
	req, err := http.NewRequest("GET", apiURL+"/s3/list", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println(string(body))
}

func getObject(bucket, key, filename string, presigned bool) {
	if filename == "" {
		filename = key
	}
	url := apiURL + "/s3/" + bucket + "/" + key
	if presigned {
		url += "?presigned=true"
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("x-api-key", apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err)
		return
	}
	defer resp.Body.Close()

	if presigned {
		var presignedURLResponse PresignedURLResponse
		if err := json.NewDecoder(resp.Body).Decode(&presignedURLResponse); err != nil {
			fmt.Println("Error decoding presigned URL response:", err)
			return
		}

		resp, err = http.Get(presignedURLResponse.URL)
		if err != nil {
			fmt.Println("Error executing presigned URL request:", err)
			return
		}
		defer resp.Body.Close()
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return
	}

	fmt.Println("File downloaded successfully")
}

func putObject(bucket, key, filename string, presigned bool) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:",
			err)
		return
	}

	url := apiURL + "/s3/" + bucket + "/" + key
	if presigned {
		req, err := http.NewRequest("GET", url+"?presigned=true", nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		req.Header.Set("x-api-key", apiKey)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error executing request:", err)
			return
		}
		defer resp.Body.Close()

		var presignedURLResponse PresignedURLResponse
		if err := json.NewDecoder(resp.Body).Decode(&presignedURLResponse); err != nil {
			fmt.Println("Error decoding presigned URL response:", err)
			return
		}

		url = presignedURLResponse.URL
	}

	req, err := http.NewRequest("PUT", url, bytes.NewReader(data))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	if !presigned {
		req.Header.Set("x-api-key", apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("File uploaded successfully")
}

func deleteObject(bucket, key string) {
	req, err := http.NewRequest("DELETE", apiURL+"/s3/"+bucket+"/"+key, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("x-api-key", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error executing request:", err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("Object deleted successfully")
}

func displayHelp() {
	fmt.Println(`Usage: your_cli_binary [OPTIONS]

OPTIONS:
 Operation to perform (list, get, put, delete)
 
 -bucket string
        Bucket name
  -key string
        Object key
  -file string
        File to upload or download
  -presigned
        Use presigned URL for get and put operations
  -help
        Display help
		
Before running the program, make sure to replace "your_api_gateway_url", "your_region", and "your_api_key" with your actual API Gateway URL, region, and API key.

You can then build and run the CLI by executing "go build" and using the generated binary. Example usage:

- List buckets: "./your_cli_binary list"
- Get object: "./your_cli_binary get -bucket my-bucket -key my-object -file output.txt -presigned true"
- Put object: "./your_cli_binary put -bucket my-bucket -key my-object -file input.txt -presigned true"
- Delete object: "./your_cli_binary delete -bucket my-bucket -key my-object"`)
}
