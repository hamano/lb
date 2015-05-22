package main

import (
	"fmt"
	"log"
	"github.com/codegangsta/cli"
	"github.com/mqu/openldap"
)

func setupBase(c *cli.Context) {
	baseDN := c.String("b")
	fmt.Printf("Adding base entry: %s\n", baseDN)
	ldap, err := openldap.Initialize(c.Args().First())
	if err != nil {
		log.Fatal("initialize error: ", err)
	}
	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
	err = ldap.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
	}
	attrs := map[string][]string{
		"objectClass": {"dcObject", "organization"},
		"dc": {"example"},
		"o": {"example"},
	}
	err = ldap.Add(baseDN, attrs)
	if err != nil {
		log.Fatal("add error: ", err)
	}
	fmt.Printf("Added base entry: %s\n", baseDN)
}
