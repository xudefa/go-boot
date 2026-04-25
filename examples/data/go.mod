module examples/data

go 1.25.0

replace (
	github.com/xudefa/go-boot/data => ../../data
	github.com/xudefa/go-boot/data/gorm => ../../data/gorm
	github.com/xudefa/go-boot/data/xorm => ../../data/xorm
)

require (
	github.com/xudefa/go-boot/data v0.0.0
	github.com/xudefa/go-boot/data/gorm v0.0.0
	github.com/xudefa/go-boot/data/xorm v0.0.0
)

require (
	filippo.io/edwards25519 v1.1.0 // indirect
	github.com/go-sql-driver/mysql v1.8.1 // indirect
	github.com/goccy/go-json v0.10.5 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	gorm.io/driver/mysql v1.5.7 // indirect
	gorm.io/gorm v1.25.12 // indirect
	xorm.io/builder v0.3.13 // indirect
	xorm.io/xorm v1.3.11 // indirect
)
