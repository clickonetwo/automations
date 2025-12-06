import {
    GbTransactionData,
    GbListTransactionsPayload,
    GbPlanData,
    GbListPlansPayload,
    GbCampaignData,
    GbListCampaignsPayload,
} from "./payloads";

export async function fetchAllGbTransactions() {
    const gbApiKey = "8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs";
    let gbApiEndpoint = `https://api.givebutter.com/v1/transactions`;
    let page = 1;
    let more = true;
    const donations: GbTransactionData[] = [];
    while (more) {
        const response = await fetch(gbApiEndpoint + `?page=${page}`, {
            method: "GET",
            headers: {
                Accept: "application/json",
                Authorization: `Bearer ${gbApiKey}`,
            },
        });
        if (!response.ok) {
            throw new Error(
                `Fetch error on page ${page}: ${response.status}, ${response.statusText}`
            );
        }
        page++;
        const body: GbListTransactionsPayload = await response.json();
        const last_page = body.meta.last_page ?? 0;
        more = last_page > page;
        for (const payload of body.data) {
            donations.push(payload);
        }
        console.log(`After page ${page} of ${last_page}, found ${donations.length} donations`);
    }
    return donations;
}

export async function fetchAllGbPlans() {
    const gbApiKey = "8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs";
    let gbApiEndpoint = `https://api.givebutter.com/v1/plans`;
    let page = 1;
    let more = true;
    const plans: GbPlanData[] = [];
    while (more) {
        const response = await fetch(gbApiEndpoint + `?page=${page}`, {
            method: "GET",
            headers: {
                Accept: "application/json",
                Authorization: `Bearer ${gbApiKey}`,
            },
        });
        if (!response.ok) {
            throw new Error(
                `Fetch error on page ${page}: ${response.status}, ${response.statusText}`
            );
        }
        page++;
        const body: GbListPlansPayload = await response.json();
        const last_page = body.meta.last_page ?? 0;
        more = last_page > page;
        for (const payload of body.data) {
            plans.push(payload);
        }
        console.log(`After page ${page} of ${last_page}, found ${plans.length} plans`);
    }
    return plans;
}
export async function fetchAllGbCampaigns() {
    const gbApiKey = "8513|QykGq6xF69yvSDsWsJG4fGq6OsrvLRwrG4TvW5vs";
    let gbApiEndpoint = `https://api.givebutter.com/v1/campaigns`;
    let page = 1;
    let more = true;
    const campaigns: GbCampaignData[] = [];
    while (more) {
        const response = await fetch(gbApiEndpoint + `?page=${page}`, {
            method: "GET",
            headers: {
                Accept: "application/json",
                Authorization: `Bearer ${gbApiKey}`,
            },
        });
        if (!response.ok) {
            throw new Error(
                `Fetch error on page ${page}: ${response.status}, ${response.statusText}`
            );
        }
        page++;
        const body: GbListCampaignsPayload = await response.json();
        const last_page = body.meta.last_page ?? 0;
        more = last_page > page;
        for (const payload of body.data) {
            campaigns.push(payload);
        }
        console.log(`After page ${page} of ${last_page}, found ${campaigns.length} campaigns`);
    }
    return campaigns;
}
