package main

import (
	"flag"
	"github.com/zero-boilerplate/go-api-helpers/service"
	"log"

	"github.com/golang-devops/go-psexec/shared"
)

const (
	CURRENT_USER_VAL = "use_current"
)

var (
	//These flags is not run as service but will exit after completion
	genpemFlag        = flag.String("genpem", "", "The full path where to generate the pem file containing the private (and public) key")
	genpubFromPemFlag = flag.String("pub_from_pem", "", "Generate the public key from the input pem file")
)

var (
	serviceUsernameFlag       = flag.String("service_username", "", "The username of the installed service (use '"+CURRENT_USER_VAL+"' without quotes to use the current user running the install service command.")
	servicePasswordFlag       = flag.String("service_password", "", "The password of the installed service")
	addressFlag               = flag.String("address", ":62677", "The full host and port to listen on")
	allowedPublicKeysFileFlag = flag.String("allowed_public_keys_file", "", "The path to the allowed public keys file")
	serverPemFlag             = flag.String("server_pem", "", "The file path for the server pem (private+public) key file")
)

func main() {
	flag.Parse()

	if len(*genpemFlag) > 0 {
		err := shared.GenerateKeyPairPemFile(*genpemFlag)
		if err != nil {
			log.Fatalf("Unable to generate key pair pem file, error: %s", err.Error())
		}
		return
	}

	if len(*genpubFromPemFlag) > 0 {
		err := shared.PrintPemFilePublicKeyAsHex(*genpubFromPemFlag)
		if err != nil {
			log.Fatalf("Unable to generate public key from pem file, error: %s", err.Error())
		}
		return
	}

	var additionalArgs []string = []string{}

	if len(*service.ServiceFlag) == 0 || *service.ServiceFlag != "uninstall" {
		if len(*serverPemFlag) == 0 {
			flag.Usage()
			log.Fatalln("The server pem flag is required.")
		}
		if len(*allowedPublicKeysFileFlag) == 0 {
			flag.Usage()
			log.Fatalln("No allowed public keys file specified, no keys will be allowed.")
		}

		additionalArgs = []string{
			"-address",
			*addressFlag,
			"-server_pem",
			*serverPemFlag,
			"-allowed_public_keys_file",
			*allowedPublicKeysFileFlag,
		}
	}

	a := &app{}

	builder := service.NewServiceRunnerBuilder("GoPsExec", a).WithOnStopHandler(a).WithAdditionalArguments(additionalArgs...)

	if len(*serviceUsernameFlag) > 0 {
		if *serviceUsernameFlag == CURRENT_USER_VAL {
			builder = builder.WithServiceUserName_AsCurrentUser()
		} else {
			builder = builder.WithServiceUserName(*serviceUsernameFlag)
		}
	}

	if len(*servicePasswordFlag) > 0 {
		builder = builder.WithServicePassword(*servicePasswordFlag)
	}

	builder.Run()
}
