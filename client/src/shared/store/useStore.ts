import { Context, useContext, useState } from 'react'

import { StoreLogicRoot } from './store-logic.root'

export const useStoreLogic = <T>(
	targetContext: Context<any>,
	instancesToAdd?: { [key: string]: any } | Array<{ [key: string]: any }>,
): T => {
	const context = useContext(targetContext)
	const [logic] = useState(() => {
		if (!instancesToAdd) return context

		return context.add(instancesToAdd)
	})

	return logic
}

export const InitStoreLogic = <T>(storeObjects: { [key: string]: any }[]) => {
	const [logic] = useState(() => {
		const root = new StoreLogicRoot()

		return root.add(storeObjects) as unknown as T
	})

	return logic as T
}
