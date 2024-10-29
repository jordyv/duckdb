<!-- PROJECT LOGO -->
<br />
<div align="center">
  <picture>
    <source media="(prefers-color-scheme: light)" srcset="https://raw.githubusercontent.com/duckdb/duckdb/main/logo/DuckDB_Logo-horizontal-dark-mode.svg">
    <source media="(prefers-color-scheme: dark)" srcset="https://raw.githubusercontent.com/duckdb/duckdb/main/logo/DuckDB_Logo-horizontal-dark-mode.svg">
    <img alt="DuckDB logo" src="https://raw.githubusercontent.com/duckdb/duckdb/main/logo/DuckDB_Logo-horizontal-dark-mode.svg" height="100">
  </picture>

<h3 align="center">GORM DuckDB Driver</h3>

  <p align="center">
    <a href="https://github.com/alifiroozi80/duckdb/issues">Report Bug</a>
    <!-- · -->
    <!-- <a href="https://github.com/alifiroozi80/duckdb/issues">Request Feature</a> -->
  </p>
</div>

---

#### 🚧 This repo is under **heavy development** and is considered non-stable. It should only be used in production once it becomes stable. 🚨

## Quick Start

```go
import (
  "github.com/alifiroozi80/duckdb"
  "gorm.io/gorm"
)

type Product struct {
	gorm.Model
	Code  string
	Price uint
}

func main() {
	// duckdb extentions: .ddb, .duckdb, .db
        db, err := gorm.Open(duckdb.Open("duckdb.ddb"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&Product{})

	// Create
	db.Create(&Product{Code: "D42", Price: 100})
}
```

Checkout [https://gorm.io](https://gorm.io) for details.

<!-- CONTRIBUTING -->

## Contributing

Any contributions you make are **greatly appreciated**.

If you have a suggestion to improve this, please fork the repo and create a pull request. You can also open an issue
with the tag "enhancement."

1) Fork the Project
2) Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3) Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4) Push to the Branch (`git push origin feature/AmazingFeature`)
5) Open a Pull Request

<!-- LICENSE -->

## License

The license is under the MIT License. See [LICENSE](https://github.com/alifiroozi80/duckdb/blob/main/LICENSE) for more
information.

## ❤ Show your support

Give a ⭐️ if this project helped you!
