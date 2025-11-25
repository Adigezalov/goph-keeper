import { useTranslation } from 'react-i18next'

import { Button } from '@shared/uikit/button'
import { InputPassword } from '@shared/uikit/input-password'
import { InputText } from '@shared/uikit/input-text'
import { Modal } from '@shared/uikit/modal'

import { TConflictResolution } from '../../types'

import styles from './conflict-resolution-modal.module.sass'

type Props = {
	visible: boolean
	onHide: () => void
	onResolve: (choice: 'local' | 'server') => void
	conflict: TConflictResolution
	isResolving: boolean
	decryptedData: {
		localLogin: string
		localPassword: string
		localBinaryData?: Uint8Array
		serverLogin: string
		serverPassword: string
		serverBinaryData?: Uint8Array
	} | null
	isLoadingDecrypted: boolean
	currentIndex: number
	totalCount: number
	onDownloadLocal: () => void
	onDownloadServer: () => void
	onNext: () => void
	onPrev: () => void
	canGoToNext: boolean
	canGoToPrev: boolean
}

export const ConflictResolutionModalView = ({
	visible,
	onHide,
	onResolve,
	conflict,
	isResolving,
	decryptedData,
	isLoadingDecrypted,
	currentIndex,
	totalCount,
	onDownloadLocal,
	onDownloadServer,
	onNext,
	onPrev,
	canGoToNext,
	canGoToPrev,
}: Props) => {
	const { t } = useTranslation()

	const formatDate = (timestamp: number) => {
		return new Date(timestamp).toLocaleString('ru-RU')
	}

	const formatDateFromString = (dateString: string) => {
		return new Date(dateString).toLocaleString('ru-RU')
	}

	const formatFileSize = (bytes: number): string => {
		if (bytes < 1024) return bytes + ' B'
		if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
		return (bytes / (1024 * 1024)).toFixed(2) + ' MB'
	}

	const getFileName = (
		metadata?: Record<string, string>,
		fallbackId?: string,
	): string => {
		if (metadata?.fileName) {
			const extension = metadata.fileExtension ? `.${metadata.fileExtension}` : ''
			return `${metadata.fileName}${extension}`
		}
		return fallbackId ? `file-${fallbackId}` : 'file'
	}

	const hasLocalBinaryData =
		conflict.localVersion.binaryData && conflict.localVersion.binaryData.length > 0
	const hasServerBinaryData =
		(conflict.serverBinaryData && conflict.serverBinaryData.length > 0) ||
		(conflict.serverVersion.binary_data_size &&
			conflict.serverVersion.binary_data_size > 0)

	const getLocalFileSize = (): string | null => {
		if (conflict.localVersion.binaryData) {
			return formatFileSize(conflict.localVersion.binaryData.length)
		}
		return null
	}

	const getServerFileSize = (): string | null => {
		if (conflict.serverBinaryData) {
			return formatFileSize(conflict.serverBinaryData.length)
		}
		if (conflict.serverVersion.binary_data_size) {
			return formatFileSize(conflict.serverVersion.binary_data_size)
		}
		return null
	}

	return (
		<Modal
			visible={visible}
			onHide={onHide}
			header={t('secrets.resolve_conflict')}
			className={styles.root}
		>
			<div className={styles.content}>
				{totalCount > 1 && (
					<div className={styles.progress}>
						<div className={styles.progress_content}>
							<Button
								icon={<i className="pi pi-chevron-left" />}
								onClick={onPrev}
								disabled={!canGoToPrev || isResolving}
								severity="secondary"
								className={styles.nav_button}
							/>
							<span className={styles.progress_text}>
								{t('secrets.conflict_progress', {
									current: currentIndex,
									total: totalCount,
								})}
							</span>
							<Button
								icon={<i className="pi pi-chevron-right" />}
								onClick={onNext}
								disabled={!canGoToNext || isResolving}
								severity="secondary"
								className={styles.nav_button}
							/>
						</div>
					</div>
				)}
				<p className={styles.description}>{t('secrets.conflict_description')}</p>

				<div className={styles.versions}>
					<div className={styles.version}>
						<div className={styles.version_header}>
							<h3 className={styles.version_title}>{t('secrets.local_version')}</h3>
							<span className={styles.version_info}>
								{t('secrets.version')}: {conflict.localVersion.version} |{' '}
								{formatDate(conflict.localVersion.updatedAt)}
							</span>
						</div>
						<div className={styles.version_content}>
							<InputText
								label={t('app')}
								value={conflict.localVersion.metadata?.app || ''}
								readOnly
								disabled
							/>
							<InputText
								label={t('username')}
								value={
									isLoadingDecrypted
										? t('secrets.loading')
										: decryptedData?.localLogin || conflict.localVersion.login
								}
								readOnly
								disabled
							/>
							<InputPassword
								label={t('password')}
								value={
									isLoadingDecrypted
										? ''
										: decryptedData?.localPassword || conflict.localVersion.password
								}
								readOnly
								disabled
							/>
							{hasLocalBinaryData && (
								<div className={styles.binary_field}>
									<label className={styles.binary_label}>
										{t('secrets.select_file')}
									</label>
									<div className={styles.binary_content}>
										<span className={styles.binary_name}>
											{getFileName(conflict.localVersion.metadata, conflict.localId)} (
											{getLocalFileSize()})
										</span>
										<Button
											severity="info"
											icon={<i className="pi pi-download" />}
											onClick={onDownloadLocal}
											disabled={!decryptedData?.localBinaryData || isLoadingDecrypted}
											className={styles.download_button}
										/>
									</div>
								</div>
							)}
						</div>
						<Button
							label={t('secrets.use_local_version')}
							onClick={() => onResolve('local')}
							disabled={isResolving}
							className={styles.version_button}
						/>
					</div>

					<div className={styles.version}>
						<div className={styles.version_header}>
							<h3 className={styles.version_title}>{t('secrets.server_version')}</h3>
							<span className={styles.version_info}>
								{t('secrets.version')}: {conflict.serverVersion.version} |{' '}
								{formatDateFromString(conflict.serverVersion.updated_at)}
							</span>
						</div>
						<div className={styles.version_content}>
							<InputText
								label={t('app')}
								value={
									(conflict.serverVersion.metadata as Record<string, string>)?.app || ''
								}
								readOnly
								disabled
							/>
							<InputText
								label={t('username')}
								value={
									isLoadingDecrypted
										? t('secrets.loading')
										: decryptedData?.serverLogin || conflict.serverVersion.login
								}
								readOnly
								disabled
							/>
							<InputPassword
								label={t('password')}
								value={
									isLoadingDecrypted
										? ''
										: decryptedData?.serverPassword || conflict.serverVersion.password
								}
								readOnly
								disabled
							/>
							{hasServerBinaryData && (
								<div className={styles.binary_field}>
									<div className={styles.binary_content}>
										<span className={styles.binary_name}>
											{getFileName(
												conflict.serverVersion.metadata as Record<string, string>,
												conflict.serverVersion.id,
											)}{' '}
											({getServerFileSize()})
										</span>
										<Button
											severity="info"
											icon={<i className="pi pi-download" />}
											onClick={onDownloadServer}
											disabled={!decryptedData?.serverBinaryData || isLoadingDecrypted}
											className={styles.download_button}
										/>
									</div>
								</div>
							)}
						</div>
						<Button
							label={t('secrets.use_server_version')}
							onClick={() => onResolve('server')}
							disabled={isResolving}
							className={styles.version_button}
							severity="info"
						/>
					</div>
				</div>
			</div>
		</Modal>
	)
}
