package db

import (
	"database/sql"
	"fmt"
	"github.com/bakape/meguca/common"
)

// ClosePost closes an open post and commits any links and hash commands
func ClosePost(id, op uint64, body string, links []common.Link, com []common.Command, claude *common.ClaudeState) (cid uint64, err error) {
	err = InTransaction(false, func(tx *sql.Tx) (err error) {
		if claude != nil {
			err = sq.Insert("claude").
				Columns("state", "prompt", "response").
				Values("waiting", claude.Prompt, claude.Response.String()).
				Suffix("RETURNING id").
				RunWith(tx).
				QueryRow().
				Scan(&cid)
			if err != nil {
				return
			}
		}
		postsMap := map[string]interface{}{
			"editing":  false,
			"body":     body,
			"commands": commandRow(com),
			"password": nil,
		}
		if claude != nil {
			postsMap["claude_id"] = cid
		}
		_, err = sq.Update("posts").
			SetMap(postsMap).
			Where("id = ?", id).
			RunWith(tx).
			Exec()
		if err != nil {
			return
		}
		err = writeLinks(tx, id, links)
		return
	})
	if err != nil {
		return
	}

	if !common.IsTest {
		// TODO: Propagate this with DB listener
		err = common.ClosePost(id, op, links, com, claude)
		if err != nil {
			return
		}
	}

	err = deleteOpenPostBody(id)
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

		if claude.Status == common.Done || claude.Status == common.Error {
			err = deleteClaude(id)
		}
		return
	})

	return
}
