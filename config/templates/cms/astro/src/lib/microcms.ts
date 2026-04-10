import { createClient } from "microcms-js-sdk";
import { MICROCMS_SERVICE_DOMAIN, MICROCMS_API_KEY } from "astro:env/server";


export const useMicroCMSClient = () => {
  const client = createClient({
    apiKey: MICROCMS_API_KEY,
    serviceDomain: MICROCMS_SERVICE_DOMAIN,
  });

  return client;
};

