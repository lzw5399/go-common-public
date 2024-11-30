package pwdutil

import (
	"testing"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"gopkg.in/go-playground/assert.v1"
)

func TestGenClientPwdFromRawPwd(t *testing.T) {
	t.Run("sm3", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.PwdHashType = "sm3"
		rawPwd := "Abcd@1234"

		// act
		clientPwd := GenClientPwdFromRawPwd(rawPwd)

		// assert
		assert.Equal(t, clientPwd, "15a4c509d1c418080bbd24133a1b4a9c39437aff92f914a53ca38e9ef442a3d1_e0bf0c108864b0706d5ce3ee93e94e3686842d76211228b0bf54dabbf08a5ca3")
	})

	t.Run("sha256", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.PwdHashType = "sha256"
		rawPwd := "Abcd@1234"

		// act
		clientPwd := GenClientPwdFromRawPwd(rawPwd)

		// assert
		assert.Equal(t, clientPwd, "584a4723121702df147ef518c208ff402b6d692407ea8d3e48cd5c9b75fdab96_ae14967257f1feba4ef33a1dcde8148832d0ced44a39d811b55dc7e63b30ab61")
	})
}

func TestGenDBPwdFromClientPwd(t *testing.T) {
	t.Run("sm3", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.PwdHashType = "sm3"
		clientPwd := "15a4c509d1c418080bbd24133a1b4a9c39437aff92f914a53ca38e9ef442a3d1_e0bf0c108864b0706d5ce3ee93e94e3686842d76211228b0bf54dabbf08a5ca3"

		// act
		dbPwd := GenDBPwdFromClientPwd(clientPwd, "")

		// assert
		assert.Equal(t, dbPwd, "3a7b78dd5cf1c0b1664994e281ed9a38f92bc860ab94c0fa39ea5affc7426449")
	})

	t.Run("sha256", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.PwdHashType = "sha256"
		clientPwd := "584a4723121702df147ef518c208ff402b6d692407ea8d3e48cd5c9b75fdab96_ae14967257f1feba4ef33a1dcde8148832d0ced44a39d811b55dc7e63b30ab61"

		// act
		dbPwd := GenDBPwdFromClientPwd(clientPwd, "")

		// assert
		assert.Equal(t, dbPwd, "584a4723121702df147ef518c208ff402b6d692407ea8d3e48cd5c9b75fdab96")
	})
}

func TestCheckClientPwd(t *testing.T) {
	t.Run("sm3 match", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.PwdHashType = "sm3"
		clientPwd := "15a4c509d1c418080bbd24133a1b4a9c39437aff92f914a53ca38e9ef442a3d1_e0bf0c108864b0706d5ce3ee93e94e3686842d76211228b0bf54dabbf08a5ca3"
		dbPwd := "3a7b78dd5cf1c0b1664994e281ed9a38f92bc860ab94c0fa39ea5affc7426449"

		// act
		isMatch, _ := CheckClientPwd(clientPwd, dbPwd, "")

		// assert
		assert.Equal(t, isMatch, true)
	})

	t.Run("sm3 unmatch", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.PwdHashType = "sm3"
		clientPwd := "15a4c509d1c418080bbd24133a1b4a9c39437aff92f914a53ca38e9ef442a3d1_e0bf0c108864b0706d5ce3ee93e94e3686842d76211228b0bf54dabbf08a5ca3"
		dbPwd := "3a7b78dd5cf1c0b1664994e281ed9a38f92bc860ab94c0fa39ea5affc74264491"

		// act
		isMatch, _ := CheckClientPwd(clientPwd, dbPwd, "")

		// assert
		assert.Equal(t, isMatch, false)
	})

	t.Run("sha256 match", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.PwdHashType = "sha256"
		clientPwd := "584a4723121702df147ef518c208ff402b6d692407ea8d3e48cd5c9b75fdab96_ae14967257f1feba4ef33a1dcde8148832d0ced44a39d811b55dc7e63b30ab61"
		dbPwd := "584a4723121702df147ef518c208ff402b6d692407ea8d3e48cd5c9b75fdab96"

		// act
		isMatch, _ := CheckClientPwd(clientPwd, dbPwd, "")

		// assert
		assert.Equal(t, isMatch, true)
	})

	t.Run("sha256 unmatch", func(t *testing.T) {
		// arrange
		fconfig.DefaultConfig.PwdHashType = "sha256"
		clientPwd := "584a4723121702df147ef518c208ff402b6d692407ea8d3e48cd5c9b75fdab96_ae14967257f1feba4ef33a1dcde8148832d0ced44a39d811b55dc7e63b30ab61"
		dbPwd := "584a4723121702df147ef518c208ff402b6d692407ea8d3e48cd5c9b75fdab92"

		// act
		isMatch, _ := CheckClientPwd(clientPwd, dbPwd, "")

		// assert
		assert.Equal(t, isMatch, false)
	})
}
