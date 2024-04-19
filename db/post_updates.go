package db

import (
	"database/sql"
	"fmt"
	"github.com/bakape/meguca/common"
	"log"
	"time"
)

// ClosePost closes an open post and commits any links and hash commands
func ClosePost(id, op uint64, body string, links []common.Link, com []common.Command, claude *common.ClaudeState) (cid uint64, err error) {
	funcStart := time.Now()

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
		}

		query := sq.Update("posts").
			Set("editing", false).
			Set("body", body).
			Set("commands", commandRow(com)).
			Set("password", nil)

		if claude != nil {
			query = query.Set("claude_id", cid)
		}

		start = time.Now()
		_, err = query.
			Where("id = ?", id).
			RunWith(tx).
			Exec()
		log.Printf("Updating posts table took %v", time.Since(start))
		if err != nil {
			return
		}

		start = time.Now()
		err = writeLinks(tx, id, links)
		log.Printf("writeLinks took %v", time.Since(start))
		return
	})
	log.Printf("InTransaction took %v", time.Since(funcStart))

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
	_ = InTransaction(false, func(tx *sql.Tx) (err error) {
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

	return
}
