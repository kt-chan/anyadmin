package main

import (
	"anyadmin-backend/pkg/utils"
	"fmt"
	"log"
)

func main() {
	enc := "ZfiYb+yjRUJEVobULjXovutPWn8iGHCriFcyhfzOewgSXaEtOx1qw9B1exlKYbPnY4+SfpotBR6LzhtMSeBKZx3eGsNQNHvLHPIu9LJ1Im3N2KCW4tuXXHRJaxJd0qA1NUswQRqvbwbShtbeBFcX+Dr7bnrAmN7l6XQcExMnBySOcpEgdtpYxdMNn+qrc43F5htaAT5NoeyuaSFd6qWwEab+fTzyS1dWwqvqYtrz+zazFp2rMpobrlShB1L2YlotVC0caLNkDE18dK3gJes7x3gVcsN8WtWA4UZr+/W4jxQxosnqWU1R1E5AMR+9a9lyb555TQMtETTP60axPerskOS6F02n4vlHt1IoHa6L1exr4qvFoMAjA5CjA+EBSEDnMgq76fJtVdMzdCMag4+uRnpLMNpsiXHPI4XDnRSJSjFw0a1VzYGq6JDshlWOC6dhWnHBtOhMBzgpacnYQgcmamdZ33EDWu7w2LqS3TVvAijYlOG5TjuVtFhJAFvnAv8lJZkBzYxVWqS15e1K4Of8Bo/B2MUnry+EFoGk9woJqpdljg/wBaDJHEc3RvNSQUBAZGAi/mKQYIlLerTNn/yOlsV4Z7cb0YSZc6F4Ej2Fgryd+jJgOjHyw2nvMisiqwPUh+1lNe9+VraFjuBibuYffvqYw4wXAN1tIb4uXAUOZJY="
	
	dec, err := utils.DecryptPassword(enc)
	if err != nil {
		log.Fatalf("Decryption failed: %v", err)
	}
	
	fmt.Printf("Decrypted: %s\n", dec)
}
