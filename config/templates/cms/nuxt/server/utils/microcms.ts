import { createClient } from "microcms-js-sdk";

export const useMicroCMSClient = () => {
  const { microCMS } = useRuntimeConfig();
  const client = createClient({
    apiKey: microCMS.apiKey,
    serviceDomain: microCMS.serviceDomain,
  });

  return client;
};

