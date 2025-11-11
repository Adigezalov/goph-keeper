import { AuthStore } from '@entities/auth/models'
import { CryptoStore } from '@entities/crypto/models'

export type TStoreLogic = {
	auth: AuthStore
	cryptoKey: CryptoStore
}
