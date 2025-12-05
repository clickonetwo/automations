import { GbTransactionData, GbListDonationsPayload } from "./payloads";

export async function fetchAllGbDonations() {
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
        const body: GbListDonationsPayload = await response.json();
        const last_page = body.meta.last_page ?? 0;
        more = last_page > page;
        for (const payload of body.data) {
            donations.push(payload);
        }
        console.log(`After page ${page} of ${last_page}, found ${donations.length} donations`);
    }
    return donations;
}
