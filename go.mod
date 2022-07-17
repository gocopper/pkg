module github.com/gocopper/pkg

go 1.16

replace github.com/gocopper/copper => ../copper

require (
	github.com/PuerkitoBio/goquery v1.8.0
	github.com/aws/aws-sdk-go v1.40.32
	github.com/gocopper/copper v0.6.1
	github.com/google/uuid v1.3.0
	github.com/google/wire v0.5.0
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	gorm.io/driver/sqlite v1.3.2
	gorm.io/gorm v1.23.5
)
