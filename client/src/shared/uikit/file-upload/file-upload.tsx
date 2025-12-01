import { forwardRef } from 'react'
import { FileUploadProps, FileUpload as PrimeFileUpload } from 'primereact/fileupload'

export const FileUpload = forwardRef<PrimeFileUpload, FileUploadProps>((props, ref) => {
	return <PrimeFileUpload ref={ref} {...props} customUpload mode="basic" />
})

FileUpload.displayName = 'FileUpload'
