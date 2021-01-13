package repository

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Edu15/recipe-golang-webservice/src/domain"

	// Postgres driver
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "edu"
	password = "1234"
	dbname   = "recipes_db"
)

// Interface interface
type Interface interface {
	FetchRecipe(recipeID int) (*domain.Recipe, error)
	FetchAuthor(ID int) (*domain.RecipeAuthor, error)
	FetchCategory(ID int) (*domain.RecipeCategory, error)
	FetchDificulty(ID int) (*domain.RecipeDifficulty, error)
	FetchRecipePreviews(w http.ResponseWriter, r *http.Request) (*[]domain.RecipePreview, error)
	UpdateRecipe(w http.ResponseWriter, r *http.Request, id int) error
	InsertRecipe(w http.ResponseWriter, r *http.Request) (int, error)
	RemoveRecipe(w http.ResponseWriter, r *http.Request, id int) error
	FetchCategories() (*[]domain.RecipeCategory, error)
	FetchDifficulties() (*[]domain.RecipeDifficulty, error)
}

// Repository struct
type Repository struct {
	db *sql.DB
}

// NewRepository method
func NewRepository() Interface {
	database := connectWithDatabase()
	//defer database.Close()

	return &Repository{
		db: database,
	}
}

func connectWithDatabase() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	return db
}

func (repo Repository) FetchRecipe(recipeID int) (*domain.Recipe, error) {
	sqlStatement := `SELECT id, title, description, author_id, category_id, dificulty_id, rating,
	preparation_time, serving, ingredients, steps, access_count, image, published_date
	FROM recipe WHERE id=$1;`

	row := repo.db.QueryRow(sqlStatement, recipeID)
	var recipe domain.Recipe
	var ingredients, steps string
	var publishedDateTime time.Time

	err := row.Scan(&recipe.ID, &recipe.Title, &recipe.Description, &recipe.Author.ID,
		&recipe.Category.ID, &recipe.Dificulty.ID, &recipe.Rating, &recipe.PreparationTime,
		&recipe.Serving, &ingredients, &steps, &recipe.AccessCount,
		&recipe.ImageURL, &publishedDateTime)

	recipe.Ingredients = strings.Split(ingredients, "|")
	recipe.Steps = strings.Split(steps, "|")
	recipe.PublishedDate = formatDate(publishedDateTime)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &recipe, nil
	default:
		return nil, err
	}
}

func formatDate(input time.Time) string {
	dateStr := input.Format(time.RFC3339)[:10]
	t, _ := time.Parse("2006-01-02", dateStr)
	return t.Format("02/Jan/2006")
}

func (repo Repository) FetchAuthor(ID int) (*domain.RecipeAuthor, error) {
	var author domain.RecipeAuthor

	sqlStatement := `SELECT id, name FROM author WHERE id=$1;`
	row := repo.db.QueryRow(sqlStatement, ID)
	err := row.Scan(&author.ID, &author.Name)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &author, nil
	default:
		return nil, err
	}
}

func (repo Repository) FetchCategory(ID int) (*domain.RecipeCategory, error) {
	var category domain.RecipeCategory

	sqlStatement := `SELECT id, name FROM category WHERE id=$1;`
	row := repo.db.QueryRow(sqlStatement, ID)
	err := row.Scan(&category.ID, &category.Name)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &category, nil
	default:
		return nil, err
	}
}

func (repo Repository) FetchDificulty(ID int) (*domain.RecipeDifficulty, error) {
	var dificulty domain.RecipeDifficulty

	sqlStatement := `SELECT id, name FROM dificulty WHERE id=$1;`
	row := repo.db.QueryRow(sqlStatement, ID)
	err := row.Scan(&dificulty.ID, &dificulty.Name)

	switch err {
	case sql.ErrNoRows:
		return nil, nil
	case nil:
		return &dificulty, nil
	default:
		return nil, err
	}
}

func (repo Repository) FetchRecipePreviews(w http.ResponseWriter, r *http.Request) (*[]domain.RecipePreview, error) {
	sqlStatement := `SELECT id, title, description FROM recipe LIMIT $1;`
	rows, err := repo.db.Query(sqlStatement, 10)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	var previews []domain.RecipePreview

	for rows.Next() {
		var preview domain.RecipePreview
		if err := rows.Scan(&preview.ID, &preview.Title, &preview.Description); err != nil {
			log.Fatal(err)
		}
		previews = append(previews, preview)
	}

	return &previews, err
}

func (repo Repository) UpdateRecipe(w http.ResponseWriter, r *http.Request, id int) error {
	sqlStatement := `UPDATE recipe 
	SET title = $2, description = $3, preparation_time = $4, serving = $5, image =$6
	WHERE id = $1;`
	title := r.FormValue("title")
	description := r.FormValue("description")
	preparationTime, _ := strconv.Atoi(r.FormValue("preparation-time"))
	serving := r.FormValue("serving")
	imageURL := r.FormValue("imgURL")
	_, err := repo.db.Exec(sqlStatement, id, title, description, preparationTime, serving, imageURL)
	return err
}

func (repo Repository) InsertRecipe(w http.ResponseWriter, r *http.Request) (int, error) {
	sqlStatement := `
	INSERT INTO recipe (title, description, author_id, category_id, dificulty_id, preparation_time, serving, ingredients, steps, image)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING id;`

	title := r.FormValue("title")
	description := r.FormValue("description")
	categoryID, _ := strconv.Atoi(r.FormValue("category"))
	difficultyID, _ := strconv.Atoi(r.FormValue("difficulty"))
	preparationTime, _ := strconv.Atoi(r.FormValue("preparation-time"))
	serving := r.FormValue("serving")
	ingredients := r.FormValue("ingredients")
	steps := r.FormValue("steps")
	imageURL := r.FormValue("imgURL")
	var id int
	err := repo.db.QueryRow(sqlStatement, title, description, 2, categoryID, difficultyID, preparationTime, serving, ingredients, steps, imageURL).Scan(&id)
	fmt.Println(id)
	return id, err
}

func (repo Repository) RemoveRecipe(w http.ResponseWriter, r *http.Request, id int) error {
	sqlStatement := `DELETE FROM recipe WHERE id = $1;`
	_, err := repo.db.Exec(sqlStatement, id)
	return err
}

func (repo Repository) FetchCategories() (*[]domain.RecipeCategory, error) {
	sqlStatement := `SELECT id, name FROM category;`
	rows, err := repo.db.Query(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	var categories []domain.RecipeCategory

	for rows.Next() {
		var category domain.RecipeCategory
		if err := rows.Scan(&category.ID, &category.Name); err != nil {
			log.Fatal(err)
		}
		categories = append(categories, category)
	}

	return &categories, err
}

func (repo Repository) FetchDifficulties() (*[]domain.RecipeDifficulty, error) {
	sqlStatement := `SELECT id, name FROM dificulty;`
	rows, err := repo.db.Query(sqlStatement)
	if err != nil {
		log.Fatal(err)
	}

	defer rows.Close()
	var dificulties []domain.RecipeDifficulty

	for rows.Next() {
		var dificulty domain.RecipeDifficulty
		if err := rows.Scan(&dificulty.ID, &dificulty.Name); err != nil {
			log.Fatal(err)
		}
		dificulties = append(dificulties, dificulty)
	}

	return &dificulties, err
}
