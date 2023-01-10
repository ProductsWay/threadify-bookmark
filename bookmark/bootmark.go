package bookmark

import (
	"context"
	"fmt"
	"time"

	"encore.dev/beta/auth"
	"encore.dev/storage/sqldb"
	"encore.dev/types/uuid"
)

// insert inserts a URL into the database.
func insert(ctx context.Context, id uuid.UUID, url string, owner auth.UID, note string) error {
	createdAt := time.Now().UTC()
	_, err := sqldb.Exec(ctx, `
		INSERT INTO bookmark (id, url, owner, note, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, id, url, owner, note, createdAt)
	return err
}

type Bookmark struct {
	ID         uuid.UUID // unique ID
	OWNER      auth.UID  // owner of the bookmark
	URL        string    // url of the bookmark
	NOTE       string    // optional note
	CREATED_AT time.Time // date time of creation
}

type BookmarkParams struct {
	URL         string // the URL to bookmark
	Description string // optional description of the bookmark
}

// Bookmark a URL.
//
//encore:api auth method=POST path=/bookmark
func CreateBookmark(ctx context.Context, p *BookmarkParams) (*Bookmark, error) {
	// print current user id to console
	id, err := uuid.NewV4()
	fmt.Printf("UUIDv4: %s\n", id)
	if err != nil {
		return nil, err
	}

	user := auth.Data()
	fmt.Print("user=", user)

	// get current user id
	uid, ok := auth.UserID()
	if !ok {
		return nil, fmt.Errorf("no user id")
	}

	owner := "github-" + uid

	if err := insert(ctx, id, p.URL, owner, p.Description); err != nil {
		return nil, err
	}

	return &Bookmark{ID: id, URL: p.URL, OWNER: owner, NOTE: p.Description}, nil
}

type GetResponse struct {
	Bookmarks []*Bookmark
}

// Get retrieves the all bookmark URLs for the owner id.
// encore:api public method=GET path=/bookmark/:id
func GetBookmarks(ctx context.Context, id string) (*GetResponse, error) {
	rows, err := sqldb.Query(ctx, `
		SELECT id, url, owner, note, created_at
		FROM bookmark
		WHERE owner = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookmarks []*Bookmark
	for rows.Next() {
		var b Bookmark
		if err := rows.Scan(&b.ID, &b.URL, &b.OWNER, &b.NOTE, &b.CREATED_AT); err != nil {
			return nil, err
		}
		bookmarks = append(bookmarks, &b)
	}

	return &GetResponse{Bookmarks: bookmarks}, nil
}

// Delete deletes a bookmark.
// encore:api auth method=DELETE path=/bookmark/:id
func DeleteBookmark(ctx context.Context, id string) error {
	// get current user id
	uid, ok := auth.UserID()
	if !ok {
		return fmt.Errorf("no user id")
	}

	owner := "github-" + uid

	_, err := sqldb.Exec(ctx, `
		DELETE FROM bookmark
		WHERE id = $1
		AND owner = $2
	`, id, owner)
	return err
}
