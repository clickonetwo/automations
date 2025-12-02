declare module 'airtable/scripts/development/payloads' {
    export = Payloads
    namespace Payloads {
        interface GbAddress {
            address_1: string,
            address_2: string | null,
            city: string,
            company: string | null,
            country: string,
            state: string,
            zipcode: string,
        }

        interface GbCustomField {
            description: string | null,
            field_id: number,
            id: number,
            title: string,
            type: string,
            value: boolean | number | string,
        }

        interface GbGivingSpace {
            amount: number,
            id: number,
            message: string | null,
            name: string,
        }

        interface GbDonationPayload {
            data: GbDonationData,
            event: string,
            id: string,
        }

        interface GbDonationData {
            address: GbAddress,
            amount: number,
            attribution_data: any[],
            campaign_code: string,
            campaign_id: number,
            communication_opt_in: boolean,
            company: string | number | null,
            company_name: string | null,
            contact_id: number,
            created_at: string    // 2025-11-25T20:23:37+00:00,
            currency: string,
            custom_fields: GbCustomField[],
            dedication: string | null,
            donated: number,
            email: string,
            external_id: string | number | null,
            fair_market_value_amount: number,
            fee: number,
            fee_covered: number,
            first_name: string,
            fund_code: string | number | null,
            fund_id: string | number | null,
            giving_space: GbGivingSpace | null,
            id: string,
            last_name: string,
            member_id: number | null,
            method: string,
            number: string,
            payment_method: string,
            payout: number,
            phone: string | null,
            plan_id: number | null,
            pledge_id: string | null,
            session_id: string,
            status: string,
            tax_deductible_amount: number,
            team_id: number | null,
            timezone: string,
            transacted_at: string // 2025-11-25T20:23:37+00:00,
            transactions: any[],
            utm_parameters: any[],
        }

        interface GbListDonationsPayload {
            data: GbDonationData[],
            links: GbListLinks,
            meta: GbListMeta,
        }

        interface GbPlanData {
            id: string,
            contact_id: number,
            first_name: string,
            last_name: string,
            email: string,
            phone: string | null,
            frequency: string,
            status: string,
            method: string,
            amount: number,
            fee_covered: string,
            created_at: string,       // UTC 2025-06-21 01:24:19,
            start_at: string,         // UTC 2025-06-21 01:24:19,
            canceled_at: string | null,
            next_bill_date: string,   // UTC 2025-12-21 00:00:00
        }

        interface GbContactEmail {
            type: string,
            value: string,
        }

        interface GbContactPhone {
            type: string,
            value: string,
        }

        interface GbContactStats {
            total_contributions: number,
            recurring_contributions: number,
        }

        interface GbContactData {
            id: number,
            external_id: string | null,
            type: string,
            prefix: string | null,
            first_name: string,
            middle_name: string | null,
            last_name: string,
            suffix: string | null,
            gender: string | null,
            dob: string | null,
            company: string | null,
            employer: string | null,
            company_name: string | null,
            point_of_contact: string | null,
            associated_companies: any[],
            title: string | null,
            website_url: string | null,
            twitter_url: string| null,
            linkedin_url: string | null,
            facebook_url: string | null,
            emails: GbContactEmail[],
            phones: GbContactPhone[],
            primary_email: string,
            primary_phone: string | null,
            note: string | null,
            addresses: GbAddress[],
            primary_address: GbAddress,
            stats: GbContactStats,
            tags: string[],
            custom_fields: GbCustomField[],
            external_ids: string[],
            is_email_subscribed: boolean,
            is_phone_subscribed: boolean,
            is_address_subscribed: boolean,
            email_opt_in: boolean,
            sms_opt_in: boolean,
            address_unsubscribed_at: string | null,
            archived_at: string | null,
            created_at: string,       // 2025-01-18T01:10:55+00:00,
            updated_at: string,       // 2025-11-25T00:28:15+00:00,
            preferred_name: string | null,
            salutation_name: string | null,
        }

        interface GbListLinks {
            first: string,
            last: string,
            next: string | null,
            prev: string | null,
        }

        interface GbListMeta {
            current_page: number,
            from: number,
            last_page: number,
            links: any[],
            path: string,
            per_page: number,
            to: number,
            total: number,
        }

        interface GbListContactsPayload {
            data: GbContactData[],
            links: GbListLinks,
            meta: GbListMeta,
        }
    }
}
