package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cloudinary/cloudinary-go"
	"github.com/cloudinary/cloudinary-go/api/uploader"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF8")
		w.WriteHeader(http.StatusOK)
		
		json.NewEncoder(w).Encode("Hello Cloudinary-go")
	})
	http.HandleFunc("/apple", uploadToCloudinary)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s := &http.Server{
		Addr:           ":" + os.Getenv("PORT"),
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	go func() {
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	fmt.Printf("Running on port : %s, http://localhost:%s/ \n", os.Getenv("PORT"), os.Getenv("PORT"))

	<-ctx.Done()
	stop()
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(timeoutCtx); err != nil {
		fmt.Println(err)
	}

}

func uploadToCloudinary(w http.ResponseWriter, r *http.Request) {
	// 1. Add your Cloudinary credentials and create a context
	cld, _ := cloudinary.NewFromParams(os.Getenv("CLOUD_NAME"), os.Getenv("API_KEY"), os.Getenv("API_SECRET"))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	// 2. Upload an image
	//===================

	resp, err := cld.Upload.Upload(ctx, "apple.png", uploader.UploadParams{PublicID: "docs/sdk/go/apple",
		Transformation: "c_crop,g_center/q_auto/f_auto", Tags: []string{"fruit"}})
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	// 3. Get your image information
	//===================

	my_image, err := cld.Image("docs/sdk/go/apple")
	if err != nil {
		fmt.Println("error")
	}

	url, err := my_image.String()
	if err != nil {
		fmt.Println("error")
	}

	fmt.Printf("url: %v\n", resp.URL)

	w.Header().Set("Content-Type", "application/json; charset=UTF8")
	w.WriteHeader(200)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"url": url,
	})

}
