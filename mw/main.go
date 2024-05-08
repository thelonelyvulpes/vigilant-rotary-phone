package main

import (
	"context"
	"flag"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"log"
	"os"
	"runtime/pprof"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()

	ctx := context.Background()
	dbUri := "bolt://localhost:7687"
	dbUser := "neo4j"
	dbPassword := "password"
	driver, err := neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(ctx)
	_, err = session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			result, _ := tx.Run(ctx, `
				UNWIND RANGE(1, 100000) as n
				RETURN n
            `, map[string]any{})
			records, _ := result.Collect(ctx)
			return records, nil
		})
	if err != nil {
		panic(err)
	}

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(*cpuprofile)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	_, err = session.ExecuteRead(ctx,
		func(tx neo4j.ManagedTransaction) (any, error) {
			pprof.StartCPUProfile(f)
			result, _ := tx.Run(ctx, `
				UNWIND RANGE(1, 100000) as n
				RETURN n
            `, map[string]any{})
			records, _ := result.Collect(ctx)
			pprof.StopCPUProfile()
			return records, nil
		})
}
