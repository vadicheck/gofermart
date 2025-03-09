package constants

type headerKey string

const XUserID headerKey = "X-User-ID"

type OrderStatus string

const StatusNew OrderStatus = "NEW"
const StatusProcessing OrderStatus = "PROCESSING"
const StatusInvalid OrderStatus = "INVALID"
const StatusProcessed OrderStatus = "PROCESSED"
