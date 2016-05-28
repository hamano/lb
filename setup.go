package main

import (
	"fmt"
	"log"
	"strings"
	"strconv"
	"github.com/urfave/cli"
	openldap "github.com/hamano/golang-openldap"
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
	if ! c.Bool("q") {
		fmt.Printf("Adding base entry: %s\n", baseDN)
	}
	err = ldap.Add(baseDN, attrs)
	if err != nil {
		log.Fatal("add error: ", err)
	}
	if ! c.Bool("q") {
		fmt.Printf("Added base entry: %s\n", baseDN)
	}
	ldap.Close()
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
		Value: "secret",
		Usage: "userPassword attribute",
	},
	cli.IntFlag {
		Name: "first",
		Value: 1,
		Usage: "first id",
	},
	cli.IntFlag {
		Name: "last",
		Value: 0,
		Usage: "last id",
	},
}

func setupPerson(c *cli.Context) {
	ldap, err := openldap.Initialize(c.Args().First())
	if err != nil {
		log.Fatal("initialize error: ", err)
	}
	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)
	err = ldap.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
	}
	last := c.Int("last")

	if last > 0 {
		for i := c.Int("first"); i <= last; i++ {
			var cn string
			if strings.Contains(c.String("cn"), "%") {
				cn = fmt.Sprintf(c.String("cn"), i)
			}else {
				cn = c.String("cn") + strconv.Itoa(i)
			}
			setupPersonOne(c, ldap, cn)
		}
	}else{
		setupPersonOne(c, ldap, c.String("cn"))
	}
	ldap.Close()
}

func setupPersonOne(c *cli.Context, ldap *openldap.Ldap, cn string) {
	baseDN := c.String("b")
	sn := c.String("sn")
	if sn == "" {
		sn = cn
	}
	userPassword := c.String("userpassword")
	dn := fmt.Sprintf("cn=%s,%s", cn, baseDN)
	attrs := map[string][]string{
		"objectClass": {"person"},
		"cn": {cn},
		"sn": {sn},
		"userPassword": {userPassword},
	}
	if ! c.Bool("q") {
		fmt.Printf("Adding person entry: %s\n", dn)
	}
	err := ldap.Add(dn, attrs)
	if err != nil {
		log.Fatal("add error: ", err)
	}
	if ! c.Bool("q") {
		fmt.Printf("Added person entry: %s\n", dn)
	}
}
