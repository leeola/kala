package main

import (
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/leeola/fixity"
	"github.com/leeola/fixity/util/dyntabwriter"
	"github.com/urfave/cli"
)

const (
	blocktypeContent = "Content: "
)

func BlocksCmd(ctx *cli.Context) error {
	fixi, err := loadFixity(ctx)
	if err != nil {
		return err
	}

	w := dyntabwriter.New(os.Stdout)
	defer w.Flush()
	w.Header(" ", "BLOCK", "HASH", "TYPE", "CONTENT", "ID")

	b, err := fixi.Blockchain().Head()
	if err == fixity.ErrEmptyBlockchain {
		return nil
	}
	if err != nil {
		return err
	}

	c, err := b.Content()
	if err != nil {
		return err
	}

	showBlockHashes := ctx.Bool("block-hashes")
	showContentHashes := ctx.Bool("content-hashes")

	bHash := sumHash(b.Hash, showBlockHashes)
	cHash := sumHash(b.Hash, showContentHashes)

	w.Println(" ",
		color.GreenString(strconv.Itoa(b.Block)),
		color.GreenString(bHash),
		color.GreenString(blockType(b)),
		color.YellowString(cHash),
		color.YellowString(c.Id),
	)

	for i := 0; i < ctx.Int("limit")-1 && b.PreviousBlockHash != ""; i++ {
		b, err = b.PreviousBlock()
		if err != nil {
			return err
		}

		c, err = b.Content()
		if err != nil {
			return err
		}

		bHash := sumHash(b.Hash, showBlockHashes)
		cHash := sumHash(b.Hash, showContentHashes)

		w.Println(" ",
			color.GreenString(strconv.Itoa(b.Block)),
			color.GreenString(bHash),
			color.GreenString(blockType(b)),
			color.YellowString(cHash),
			color.YellowString(c.Id),
		)
	}

	return nil
}

func sumHash(h string, doNothing bool) string {
	if doNothing {
		return h
	}
	return h[len(h)-8:]
}

func blockType(b fixity.Block) string {
	switch {
	case b.ContentHash != "":
		return "content"
	default:
		return "unknown"
	}
}