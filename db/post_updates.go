package db

import (
	"database/sql"
	"fmt"
	"github.com/bakape/meguca/common"
	"github.com/lib/pq"
	"log"
	"time"
)

var (
	updatePostsStmt     *sql.Stmt
	updatePostsAndLinks *sql.Stmt
)

func prepareUpdatePostsStmt() (err error) {
	updatePostsStmt, err = sqlDB.Prepare(`
        UPDATE posts
        SET editing = $1,
            body = $2,
            commands = $3,
            password = $4,
            claude_id = $5
        WHERE id = $6
		`)
	updatePostsAndLinks, err = sqlDB.Prepare(`SELECT close_post($1, $2, $3, $4, $5)`)
	if err != nil {
		log.Fatal(err)
	}
	return
}

// ClosePost closes an open post and commits any links and hash commands
func ClosePost(id, op uint64, body string, links []common.Link, com []common.Command, claude *common.ClaudeState) (cid uint64, err error) {
	funcStart := time.Now()
	// Hotpath for closing posts without links or Claude
	if len(links) == 0 && claude == nil {
		start := time.Now()
		_, err = updatePostsStmt.Exec(false, body, commandRow(com), nil, nil, id)
		log.Printf("updatePostsStmt.Exec took %v", time.Since(start))
		if err != nil {
			return
		}
	} else if claude == nil {
		linksArray := make([]int64, len(links))
		for i, link := range links {
			linksArray[i] = int64(link.ID)
		}
		start := time.Now()
		_, err = updatePostsAndLinks.Exec(body, commandRow(com), id, nil, pq.Array(linksArray))
		log.Printf("updatePostsAndLinks.Exec took %v", time.Since(start))
	} else {
		err = InTransaction(false, func(tx *sql.Tx) (err error) {
			start := time.Now()
			if claude != nil {
				err = sq.Insert("claude").
					Columns("state", "prompt", "response").
					Values("waiting", claude.Prompt, claude.Response.String()).
					Suffix("RETURNING id").
					RunWith(tx).
					QueryRow().
					Scan(&cid)
				log.Printf("Inserting into claude table took %v", time.Since(start))
				if err != nil {
					return
				}
				_, err = tx.Stmt(updatePostsStmt).Exec(false, body, commandRow(com), nil, cid, id)
				if err != nil {
					return
				}
			} else {
				_, err = tx.Stmt(updatePostsStmt).Exec(false, body, commandRow(com), nil, nil, id)
				if err != nil {
					return
				}
			}

			err = writeLinks(tx, id, links)
			return
		})
		log.Printf("ClosePost transaction took %v", time.Since(funcStart))
	}

	if err != nil {
		return
	}

	if !common.IsTest {
		start := time.Now()
		// TODO: Propagate this with DB listener
		err = common.ClosePost(id, op, links, com, claude)
		log.Printf("common.ClosePost took %v", time.Since(start))
		if err != nil {
			return
		}
	}

	start := time.Now()
	err = deleteOpenPostBody(id)
	log.Printf("deleteOpenPostBody took %v", time.Since(start))
	log.Printf("ClosePost took %v", time.Since(funcStart))
	return
}
func UpdateClaude(id uint64, claude *common.ClaudeState) {
	err := InTransaction(false, func(tx *sql.Tx) (err error) {
		// Update the Claude associated with the post using a subquery
		result, err := sq.Update("claude").
			SetMap(map[string]interface{}{
				"state":    claude.GetStatusString(),
				"prompt":   claude.Prompt,
				"response": claude.Response.String(),
			}).
			Where("id = ?", id).
			RunWith(tx).
			Exec()
		if err != nil {
			return
		}

		// Check if any rows were affected by the update
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return
		}

		if rowsAffected == 0 {
			// If no rows were affected, the post doesn't have an associated Claude
			return fmt.Errorf("post with ID %d has no associated Claude", id)
		}

		return
	})
	if err != nil {
		log.Println(err)
	}

	return
}
