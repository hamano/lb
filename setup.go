package main

import (
	"fmt"
	"github.com/urfave/cli"
	"gopkg.in/ldap.v3"
	"log"
	"strconv"
	"strings"
)

func setupBase(c *cli.Context) error {
	baseDN := c.String("b")
	conn, err := ldap.DialURL(c.Args().First())
	if err != nil {
		log.Fatal("initialize error: ", err)
	}

	err = conn.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
	}

	req := ldap.AddRequest{
		DN: baseDN,
		Attributes: []ldap.Attribute{
			ldap.Attribute{"objectClass",
				[]string{"dcObject", "organization"}},
			ldap.Attribute{"o", []string{"lb"}},
		},
	}

	if !c.Bool("q") {
		fmt.Printf("Adding base entry: %s\n", baseDN)
	}

	err = conn.Add(&req)
	if err != nil {
		log.Fatal("add error: ", err)
	}

	if !c.Bool("q") {
		fmt.Printf("Added base entry: %s\n", baseDN)
	}

	conn.Close()
	return nil
}

var setupPersonFlags = []cli.Flag{
	&cli.StringFlag{
		Name:  "cn",
		Value: "user",
		Usage: "cn attribute",
	},
	&cli.StringFlag{
		Name:  "sn",
		Value: "",
		Usage: "sn attribute",
	},
	&cli.StringFlag{
		Name:  "password, userpassword, userPassword",
		Value: "secret",
		Usage: "userPassword attribute",
	},
	&cli.IntFlag{
		Name:  "first",
		Value: 1,
		Usage: "first id",
	},
	&cli.IntFlag{
		Name:  "last",
		Value: 0,
		Usage: "last id",
	},
}

func setupPerson(c *cli.Context) error {
	conn, err := ldap.DialURL(c.Args().First())
	if err != nil {
		log.Fatal("initialize error: ", err)
	}

	err = conn.Bind(c.String("D"), c.String("w"))
	if err != nil {
		log.Fatal("bind error: ", err)
	}

	first := c.Int("first")
	last := c.Int("last")

	if last > 0 {
		for i := first; i <= last; i++ {
			var cn string
			if strings.Contains(c.String("cn"), "%") {
				cn = fmt.Sprintf(c.String("cn"), i)
			} else {
				cn = c.String("cn") + strconv.Itoa(i)
			}
			setupOnePerson(c, conn, cn)
		}
	} else {
		setupOnePerson(c, conn, c.String("cn"))
	}
	conn.Close()
	return nil
}

func setupOnePerson(c *cli.Context, conn *ldap.Conn, cn string) error {
	baseDN := c.String("b")
	sn := c.String("sn")
	if sn == "" {
		sn = cn
	}
	dn := fmt.Sprintf("cn=%s,%s", cn, baseDN)
	userPassword := c.String("userpassword")
	req := ldap.AddRequest{
		DN: dn,
		Attributes: []ldap.Attribute{
			ldap.Attribute{"objectClass", []string{"person"}},
			ldap.Attribute{"cn", []string{cn}},
			ldap.Attribute{"sn", []string{sn}},
			ldap.Attribute{"userPassword", []string{userPassword}},
		},
	}

	if !c.Bool("q") {
		fmt.Printf("Adding person entry: %s\n", dn)
	}

	err := conn.Add(&req)
	if err != nil {
		log.Fatal("add error: ", err)
	}

	if !c.Bool("q") {
		fmt.Printf("Added person entry: %s\n", dn)
	}
	return nil
}
