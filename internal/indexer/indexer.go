package indexer

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"memento/internal/markdown"
	sqlite "memento/internal/sqlite"
)

func Index(wikiDir string, dbPath string) error {
	absWiki, err := filepath.Abs(wikiDir)
	if err != nil {
		return fmt.Errorf("resolve wiki dir: %w", err)
	}

	db, err := sqlite.OpenDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := sqlite.RecreateTables(db); err != nil {
		return err
	}

	insertSection, err := db.Prepare(`
		INSERT INTO sections (path, anchor, heading, heading_level, body, section_order)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert sections: %w", err)
	}
	defer insertSection.Close()

	insertFTS, err := db.Prepare(`
		INSERT INTO sections_fts (rowid, body, heading_weight, path, anchor, heading_level, section_order)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert fts: %w", err)
	}
	defer insertFTS.Close()

	insertLink, err := db.Prepare(`
		INSERT INTO section_links (source_section_id, target_path, target_anchor, link_text)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare insert link: %w", err)
	}
	defer insertLink.Close()

	return filepath.Walk(absWiki, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		relPath, err := filepath.Rel(absWiki, path)
		if err != nil {
			return fmt.Errorf("resolve relative path: %w", err)
		}
		relPath = filepath.ToSlash(relPath)

		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read file %s: %w", path, err)
		}

		sections, err := markdown.ParseFile(relPath, source)
		if err != nil {
			return fmt.Errorf("parse file %s: %w", path, err)
		}

		return insertSections(db, insertSection, insertFTS, insertLink, sections)
	})
}

func insertSections(db *sql.DB, insertSection, insertFTS, insertLink *sql.Stmt, sections []markdown.Section) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	for _, s := range sections {
		result, err := tx.Stmt(insertSection).Exec(s.Path, s.Anchor, s.Heading, s.HeadingLevel, s.Body, s.SectionOrder)
		if err != nil {
			return fmt.Errorf("insert section: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return fmt.Errorf("get last insert id: %w", err)
		}

		headingWeight := buildHeadingWeight(s.Heading)
		_, err = tx.Stmt(insertFTS).Exec(id, s.Body, headingWeight, s.Path, s.Anchor, s.HeadingLevel, s.SectionOrder)
		if err != nil {
			return fmt.Errorf("insert fts: %w", err)
		}

		for _, l := range s.Links {
			_, err = tx.Stmt(insertLink).Exec(id, l.TargetPath, l.TargetAnchor, l.Text)
			if err != nil {
				return fmt.Errorf("insert link: %w", err)
			}
		}
	}

	return tx.Commit()
}

func buildHeadingWeight(heading string) string {
	return heading
}
