export interface GbAddress {
    address_1: string;
    address_2: string | null;
    city: string;
    company: string | null;
    country: string; // ISO alpha-3 code
    state: string;
    zipcode: string;
}

export interface GbCustomField {
    description: string | null;
    field_id: number;
    id: number;
    title: string;
    type: string;
    value: boolean | number | string;
}

export interface GbGivingSpace {
    amount: number;
    id: number;
    message: string | null;
    name: string;
}

export interface GbTransactionPayload {
    data: GbTransactionData;
    event: string;
    id: string;
}

export interface GbTransactionData {
    address: GbAddress;
    amount: number;
    attribution_data: any[];
    campaign_code: string;
    campaign_id: number;
    communication_opt_in: boolean;
    company: string | number | null;
    company_name: string | null;
    contact_id: number;
    created_at: string; // 2025-11-25T20:23:37+00:00,
    currency: string;
    custom_fields: GbCustomField[];
    dedication: string | null;
    donated: number;
    email: string;
    external_id: string | number | null;
    fair_market_value_amount: number;
    fee: number;
    fee_covered: number;
    first_name: string;
    fund_code: string | number | null;
    fund_id: string | number | null;
    giving_space: GbGivingSpace | null;
    id: string;
    last_name: string;
    member_id: number | null;
    method: string;
    number: string;
    payment_method: string;
    payout: number;
    phone: string | null;
    plan_id: number | null;
    pledge_id: string | null;
    session_id: string;
    status: string;
    tax_deductible_amount: number;
    team_id: number | null;
    timezone: string;
    transacted_at: string; // 2025-11-25T20:23:37+00:00,
    transactions: any[];
    utm_parameters: any[];
}

export interface GbListTransactionsPayload {
    data: GbTransactionData[];
    links: GbListLinks;
    meta: GbListMeta;
}

export interface GbCampaignData {
    id: number;
    code?: string;
    account_id: string;
    event_id: number | null;
    type: string;
    title: string;
    subtitle: string | null;
    description: string | null;
    slug: string;
    url: string;
    goal: number;
    raised: number;
    donors: number;
    currency: string;
    cover: any | null;
    status: string;
    timezone?: string;
    end_at: string | null;
    event?: any;
    created_at: string; // ISO DateTime
    updated_at: string; // ISO DateTime
}

export interface GbListCampaignsPayload {
    data: GbCampaignData[];
    links: GbListLinks;
    meta: GbListMeta;
}

export interface GbPlanData {
    id: string;
    contact_id: number;
    first_name: string;
    last_name: string;
    email: string;
    phone: string | null;
    frequency: string;
    status: string;
    method: string;
    amount: number;
    fee_covered: string;
    created_at: string; // UTC 2025-06-21 01:24:19,
    start_at: string; // UTC 2025-06-21 01:24:19,
    canceled_at: string | null;
    next_bill_date: string; // UTC 2025-12-21 00:00:00
}

export interface GbListPlansPayload {
    data: GbPlanData[];
    links: GbListLinks;
    meta: GbListMeta;
}

export interface GbContactEmail {
    type: string;
    value: string;
}

export interface GbContactPhone {
    type: string;
    value: string;
}

export interface GbContactStats {
    total_contributions: number;
    recurring_contributions: number;
}

export interface GbContactData {
    id: number;
    external_id: string | null;
    type: string;
    prefix: string | null;
    first_name: string;
    middle_name: string | null;
    last_name: string;
    suffix: string | null;
    gender: string | null;
    dob: string | null;
    company: string | null;
    employer: string | null;
    company_name: string | null;
    point_of_contact: string | null;
    associated_companies: any[];
    title: string | null;
    website_url: string | null;
    twitter_url: string | null;
    linkedin_url: string | null;
    facebook_url: string | null;
    emails: GbContactEmail[];
    phones: GbContactPhone[];
    primary_email: string;
    primary_phone: string | null;
    note: string | null;
    addresses: GbAddress[];
    primary_address: GbAddress;
    stats: GbContactStats;
    tags: string[];
    custom_fields: GbCustomField[];
    external_ids: string[];
    is_email_subscribed: boolean;
    is_phone_subscribed: boolean;
    is_address_subscribed: boolean;
    email_opt_in: boolean;
    sms_opt_in: boolean;
    address_unsubscribed_at: string | null;
    archived_at: string | null;
    created_at: string; // 2025-01-18T01:10:55+00:00,
    updated_at: string; // 2025-11-25T00:28:15+00:00,
    preferred_name: string | null;
    salutation_name: string | null;
}

export interface GbListLinks {
    first: string;
    last: string;
    next: string | null;
    prev: string | null;
}

export interface GbListMeta {
    current_page: number;
    from: number;
    last_page: number;
    links: any[];
    path: string;
    per_page: number;
    to: number;
    total: number;
}

export interface GbListContactsPayload {
    data: GbContactData[];
    links: GbListLinks;
    meta: GbListMeta;
}
