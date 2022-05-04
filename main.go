package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/pianisimo/go-postgresql/models"
	"github.com/pianisimo/go-postgresql/storage"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	Db *gorm.DB
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_book/:id", r.GetBookById)
	api.Get("/books/", r.GetBooks)
}

func (r *Repository) CreateBook(ctx *fiber.Ctx) error {
	book := &Book{}
	err := ctx.BodyParser(book)
	if err != nil {
		ctx.Status(http.StatusUnprocessableEntity).JSON(&fiber.Map{"message": "request failed"})
		return err
	}

	err = r.Db.Create(book).Error
	if err != nil {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "could not save book"})
		return err
	}

	ctx.Status(http.StatusOK).JSON(&fiber.Map{"message": "book created"})
	return nil
}

func (r *Repository) DeleteBook(ctx *fiber.Ctx) error {
	book := &models.Book{}
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "id can not be empty"})
		return nil
	}

	tx := r.Db.Delete(book, id)
	if tx.Error != nil {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "could not delete book by id"})
		return tx.Error
	}

	return nil
}

func (r *Repository) GetBookById(ctx *fiber.Ctx) error {
	book := &models.Book{}
	id := ctx.Params("id")
	if id == "" {
		ctx.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "id can not be empty"})
		return nil
	}

	err := r.Db.Where("id = ?", id).First(book).Error
	if err != nil {
		ctx.Status(http.StatusNotFound).JSON(&fiber.Map{"message": "book not found"})
		return err
	}

	ctx.Status(http.StatusNotFound).JSON(&fiber.Map{"data": book})
	return nil
}

func (r *Repository) GetBooks(ctx *fiber.Ctx) error {
	books := &[]models.Book{}

	err := r.Db.Find(books).Error
	if err != nil {
		ctx.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "could not get all books"})
		return err
	}

	ctx.Status(http.StatusOK).JSON(&fiber.Map{"data": books})
	return nil
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASSWORD"),
		User:     os.Getenv("DB_USER"),
		DbName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL"),
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("could not load the database")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := Repository{
		Db: db,
	}
	app := fiber.New()
	r.SetupRoutes(app)
	err = app.Listen(":8080")
	if err != nil {
		log.Fatal(err)
		return
	}
}
