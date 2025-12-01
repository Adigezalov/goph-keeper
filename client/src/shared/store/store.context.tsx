import { createContext } from 'react'

import { TStoreLogic } from './store-logic.types'

export const StoreContextLogic = createContext<TStoreLogic | any>({})
