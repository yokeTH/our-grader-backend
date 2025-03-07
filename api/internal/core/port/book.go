package port

import (
	"github.com/gofiber/fiber/v2"
	"github.com/yokeTH/our-grader-backend/api/internal/core/domain"
)

type BookService interface {
	CreateBook(book *domain.Book) error
	GetBook(id int) (*domain.Book, error)
	GetBooks(limit int, page int) ([]*domain.Book, int, int, error)
	UpdateBook(id int, book *domain.Book) (*domain.Book, error)
	DeleteBook(id int) error
}

type BookRepository interface {
	CreateBook(book *domain.Book) error
	GetBook(id int) (*domain.Book, error)
	GetBooks(limit int, page int) ([]*domain.Book, int, int, error)
	UpdateBook(id int, book *domain.Book) (*domain.Book, error)
	DeleteBook(id int) error
}

type BookHandler interface {
	CreateBook(c *fiber.Ctx) error
	GetBook(c *fiber.Ctx) error
	GetBooks(c *fiber.Ctx) error
	UpdateBook(c *fiber.Ctx) error
	DeleteBook(c *fiber.Ctx) error
}
