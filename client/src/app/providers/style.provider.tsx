import { ReactNode } from 'react'

import '../styles/styles/main.sass'

type TProps = {
	children: ReactNode
}

export const StyleProvider = ({ children }: TProps) => children
