// Type definitions for Airtable internal scripting,
// such as automations and extensions.

declare module 'airtable_internal' {
    export = Airtable
    namespace Airtable {
        interface Input {
            config(): { [key: string]: any },
        }

        interface Base {
            getTable(idOrName: string): Table,
        }
        const base: Base
        const input: Input

        interface Table {
            id: string,
            name: string,
            description?: string,
            url: string,
            fields: Field[],
            selectRecordAsync(
                recordId: string,
                options?: {
                    fields?: Array<Field | string>,
                },
            ): Promise<Record | null>
            selectRecordsAsync(options?: {
                sorts?: Array<{
                    field: Field | string,
                    direction?: 'asc' | 'desc',
                }>,
                fields?: Array<Field | string>,
                recordIds?: Array<string>,
            }): Promise<RecordQueryResult>
            createRecordAsync(fields: {[fieldNameOrId: string]: unknown}): Promise<string>
            updateRecordAsync(
                recordOrRecordId: Record | string,
                fields: {[fieldNameOrId: string]: unknown}
            ): Promise<void>
        }

        interface Field {
            id: string,
            name: string,
            description?: string,
            type: string,
        }

        interface Record {
            id: string,
            name: string,
            getCellValue(idOrName: string): any
            getCellValueAsString(idOrName: string): string
        }

        interface RecordQueryResult {
            recordIds: string[],
            records: Record[],
        }
    }
}
