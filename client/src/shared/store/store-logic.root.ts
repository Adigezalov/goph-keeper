import { configure, makeObservable } from 'mobx'

import { StoreLogicBase } from './store-logic.base.ts'

configure({
	enforceActions: 'always',
	reactionRequiresObservable: true,
	observableRequiresReaction: true,
	computedRequiresReaction: true,
})

export class StoreLogicRoot extends StoreLogicBase {
	constructor() {
		super()
		makeObservable(this, {})
	}
}
