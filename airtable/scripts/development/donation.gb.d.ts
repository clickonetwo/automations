declare module 'donation_gb' {
    export = donation_gb
    namespace donation_gb {
        interface Payload {
            id: string,
            donated: number,
            first_name: string,
            last_name: string,
            email: string,
            start_at: string,       // ISO 8601 date
            created_at: string,     // ISO 8601 date
            canceled_at: string,    // ISO 8601 date
            contact_id: number,
            plan_id: number,
        }
    }
}
