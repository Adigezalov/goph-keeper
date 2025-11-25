import { type ComponentType, type LazyExoticComponent, ReactNode, Suspense } from 'react'

import { PageLoader } from '@shared/ui'

interface LazyLoaderProps {
	Component: LazyExoticComponent<ComponentType<any>>
	fallback?: ReactNode
}

export const LazyLoader = ({ Component, fallback }: LazyLoaderProps) => {
	return (
		<Suspense fallback={fallback || <PageLoader />}>
			<Component />
		</Suspense>
	)
}
