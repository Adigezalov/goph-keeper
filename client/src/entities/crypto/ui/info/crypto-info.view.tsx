import { classNames } from 'primereact/utils'
import { useTranslation } from 'react-i18next'

import { Icon } from '@shared/uikit/icon'
import { Tooltip } from '@shared/uikit/tooltip'

import styles from './crypto-info.module.sass'

type Props = {
	isCryptoKeySuccess: boolean
	onShowCryptoModal: () => void
	onCopyCryptoModal: () => void
}

export const CryptoInfoView = ({
	isCryptoKeySuccess,
	onShowCryptoModal,
	onCopyCryptoModal,
}: Props) => {
	const { t } = useTranslation()

	return (
		<>
			<span
				data-pr-tooltip={t(isCryptoKeySuccess ? 'copy_crypto_key' : 'no_crypto_key')}
				data-pr-position={'bottom'}
				className={classNames(
					Icon.KEY,
					styles.root,
					{ [styles.error]: !isCryptoKeySuccess },
					'crypto-key-icon',
				)}
				onClick={!isCryptoKeySuccess ? onShowCryptoModal : onCopyCryptoModal}
			/>
			<Tooltip target={'.crypto-key-icon'} />
		</>
	)
}
