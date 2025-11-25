import { Control } from 'react-hook-form'
import { useTranslation } from 'react-i18next'

import { InputTextareaField } from '@shared/reused/input-textarea-field'
import { Button } from '@shared/uikit/button'
import { Modal } from '@shared/uikit/modal'

import styles from './crypto-modal.module.sass'

type Props = {
	control: Control<{ key?: string }>
	visible: boolean
	onHide: () => void
	onSave: () => void
	onGenerate: () => void
}

export const CryptoModalView = ({
	control,
	visible,
	onHide,
	onSave,
	onGenerate,
}: Props) => {
	const { t } = useTranslation()

	return (
		<Modal visible={visible} onHide={onHide} header={t('crypto_key')}>
			<div className={styles.root}>
				<div>{t('enter_crypto_key')}</div>
				<div>{t('crypto_key_alert')}</div>
				<InputTextareaField
					control={control}
					name={'key'}
					label={t('your_crypto_key')}
					required
				/>
				<div className={styles.actions}>
					<Button
						label={t('generate_new_crypto_key')}
						outlined
						onClick={onGenerate}
						type={'button'}
					/>
					<Button label={t('save')} onClick={onSave} type={'submit'} />
				</div>
			</div>
		</Modal>
	)
}
