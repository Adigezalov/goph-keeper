import i18next from 'i18next'

function isClass(func: any) {
	return (
		typeof func === 'function' &&
		/^class\s/.test(Function.prototype.toString.call(func))
	)
}

export class StoreLogicBase {
	[key: string]: any

	add(value: object): typeof this & object {
		const addValue = (element: object) => {
			const [key, Value] = Object.entries(element)[0]

			if (this[key]) return

			if (isClass(Value)) {
				this[key] = new Value(this)
			} else {
				this[key] = Value
			}
		}

		if (typeof value === 'object') {
			if (Array.isArray(value)) {
				for (const element of value) {
					addValue(element)
				}
			} else {
				addValue(value)
			}
		} else {
			throw new Error(i18next.t('store.logic_add_error'))
		}

		return this
	}

	replace(name: string, value: object) {
		this[name] = value

		return this
	}
}
