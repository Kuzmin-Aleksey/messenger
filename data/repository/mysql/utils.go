package mysql

import (
	"context"
	"io"
	"os"
)

func openAndExec(ctx context.Context, db DB, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	script, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, string(script)); err != nil {
		return err
	}
	return nil
}
