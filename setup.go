package main

import (
	"fmt"
	"log"
	"github.com/codegangsta/cli"
	"github.com/mqu/openldap"
)

func setupBase(c *cli.Context) {
	baseDN := c.String("b")
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
		"o": {"lb"},
	}
	fmt.Printf("Adding base entry: %s\n", baseDN)
	err = ldap.Add(baseDN, attrs)
	if err != nil {
		log.Fatal("add error: ", err)
	}
	fmt.Printf("Added base entry: %s\n", baseDN)
}

var setupPersonFlags = []cli.Flag {
	cli.StringFlag {
		Name: "cn",
		Value: "user",
		Usage: "cn attribute",
	},
	cli.StringFlag {
		Name: "sn",
		Value: "",
		Usage: "sn attribute",
	},
	cli.StringFlag {
		Name: "password, userpassword, userPassword",
		Value: "password",
		Usage: "userPassword attribute",
	},
}

func setupPerson(c *cli.Context) {
	baseDN := c.String("b")
	ldap, err := openldap.Initialize(c.Args().First())
	if err != nil {
		log.Fatal("initialize error: ", err)
	}
	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
	err = ldap.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
	}
	cn := c.String("cn")
	sn := c.String("sn")
	if sn == "" {
		sn = cn
	}
	fmt.Printf("sn: %s\n", c.String("sn"))
	userPassword := c.String("userpassword")
	dn := fmt.Sprintf("cn=%s,%s", cn, baseDN)
	attrs := map[string][]string{
		"objectClass": {"person"},
		"cn": {cn},
		"sn": {sn},
		"userPassword": {userPassword},
	}
	fmt.Printf("Adding person entry: %s\n", dn)
	err = ldap.Add(dn, attrs)
	if err != nil {
		log.Fatal("add error: ", err)
	}
	fmt.Printf("Added base entry: %s\n", baseDN)
}
