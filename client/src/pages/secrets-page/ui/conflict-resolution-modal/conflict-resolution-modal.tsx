import { observer } from 'mobx-react-lite'
import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'

import { StoreContextLogic, TStoreLogic, useStoreLogic } from '@shared/store'
import { showToastNotification } from '@shared/toast-notification'
import { TOAST_SEVERITY } from '@shared/uikit/toast'

import { ConflictResolutionModalView } from './conflict-resolution-modal.view'

type DecryptedConflictData = {
	localLogin: string
	localPassword: string
	localBinaryData?: Uint8Array
	serverLogin: string
	serverPassword: string
	serverBinaryData?: Uint8Array
}

export const ConflictResolutionModal = observer(() => {
	const { t } = useTranslation()
	const store = useStoreLogic<TStoreLogic>(StoreContextLogic)

	const {
		visibleConflictResolvingModal,
		currentConflict,
		conflictsCount,
		currentConflictIndex,
		resolveConflict,
		decryptData,
		decryptBinaryData,
		goToNextConflict,
		goToPrevConflict,
		canGoToNext,
		canGoToPrev,
	} = store.secretsPage

	const [isResolving, setIsResolving] = useState(false)
	const [decryptedData, setDecryptedData] = useState<DecryptedConflictData | null>(null)
	const [isLoadingDecrypted, setIsLoadingDecrypted] = useState(false)

	useEffect(() => {
		if (
			visibleConflictResolvingModal &&
			currentConflict &&
			!decryptedData &&
			!isLoadingDecrypted
		) {
			setIsLoadingDecrypted(true)
			const loadDecrypted = async () => {
				try {
					const [
						localLogin,
						localPassword,
						localBinaryData,
						serverLogin,
						serverPassword,
						serverBinaryData,
					] = await Promise.all([
						decryptData(currentConflict.localVersion.login),
						decryptData(currentConflict.localVersion.password),
						currentConflict.localVersion.binaryData
							? decryptBinaryData(currentConflict.localVersion.binaryData)
							: Promise.resolve(undefined),
						decryptData(currentConflict.serverVersion.login),
						decryptData(currentConflict.serverVersion.password),
						currentConflict.serverBinaryData
							? decryptBinaryData(currentConflict.serverBinaryData)
							: Promise.resolve(undefined),
					])

					setDecryptedData({
						localLogin,
						localPassword,
						localBinaryData,
						serverLogin,
						serverPassword,
						serverBinaryData,
					})
				} catch (error) {
					console.error('Error decrypting conflict data:', error)
				} finally {
					setIsLoadingDecrypted(false)
				}
			}
			void loadDecrypted()
		}
	}, [
		visibleConflictResolvingModal,
		currentConflict,
		decryptedData,
		isLoadingDecrypted,
		decryptData,
		decryptBinaryData,
	])

	useEffect(() => {
		if (currentConflict) {
			setDecryptedData(null)
			setIsLoadingDecrypted(false)
		}
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, [currentConflict?.secretId, currentConflict?.localId, currentConflictIndex])

	useEffect(() => {
		if (!visibleConflictResolvingModal) {
			setIsResolving(false)
			setDecryptedData(null)
			setIsLoadingDecrypted(false)
		}
	}, [visibleConflictResolvingModal])

	const onHide = () => {
		if (!isResolving) {
			if (conflictsCount > 0) {
				showToastNotification({
					message: t('secrets.conflicts_pending_warning', { count: conflictsCount }),
					header: t('secrets.conflicts_pending'),
					severity: TOAST_SEVERITY.WARNING,
				})
			}
			store.secretsPage.setVisibleConflictResolvingModal(false)
		}
	}

	const onResolve = async (choice: 'local' | 'server') => {
		setIsResolving(true)
		try {
			await resolveConflict(choice)
			setDecryptedData(null)
		} finally {
			setIsResolving(false)
		}
	}

	const onDownloadLocal = () => {
		if (!decryptedData?.localBinaryData || !currentConflict) return

		const arrayBuffer = new Uint8Array(decryptedData.localBinaryData).buffer
		const blob = new Blob([arrayBuffer])

		const url = URL.createObjectURL(blob)
		const link = document.createElement('a')
		link.href = url

		const fileName =
			currentConflict.localVersion.metadata?.fileName || `file-${currentConflict.localId}`
		const fileExtension = currentConflict.localVersion.metadata?.fileExtension || 'bin'
		link.download = `${fileName}.${fileExtension}`

		document.body.appendChild(link)
		link.click()

		document.body.removeChild(link)
		URL.revokeObjectURL(url)
	}

	const onDownloadServer = () => {
		if (!decryptedData?.serverBinaryData || !currentConflict) return

		const arrayBuffer = new Uint8Array(decryptedData.serverBinaryData).buffer
		const blob = new Blob([arrayBuffer])

		const url = URL.createObjectURL(blob)
		const link = document.createElement('a')
		link.href = url

		const fileName =
			currentConflict.serverVersion.metadata?.fileName ||
			`file-${currentConflict.serverVersion.id}`
		const fileExtension = currentConflict.serverVersion.metadata?.fileExtension || 'bin'
		link.download = `${fileName}.${fileExtension}`

		document.body.appendChild(link)
		link.click()

		document.body.removeChild(link)
		URL.revokeObjectURL(url)
	}

	const onNext = () => {
		goToNextConflict()
	}

	const onPrev = () => {
		goToPrevConflict()
	}

	if (!currentConflict) {
		return null
	}

		return (
		<ConflictResolutionModalView
			visible={visibleConflictResolvingModal}
			onHide={onHide}
			onResolve={onResolve}
			conflict={currentConflict}
			isResolving={isResolving}
			decryptedData={decryptedData}
			isLoadingDecrypted={isLoadingDecrypted}
			currentIndex={currentConflictIndex + 1}
			totalCount={conflictsCount}
			onDownloadLocal={onDownloadLocal}
			onDownloadServer={onDownloadServer}
			onNext={onNext}
			onPrev={onPrev}
			canGoToNext={canGoToNext}
			canGoToPrev={canGoToPrev}
		/>
	)
})
