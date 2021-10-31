-- insert auth
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('armend.lesi+4@gmail.com', 'ACSKRVCANHURIBFN47N42KFO5QXGHYHF7GS6D3B6QRUVZN2VIYHTHNHYLLVGO7UOG55DNKPUPQNVNDQ=', 'PEEM5ZGXUGMUDSUVYDU5V4EGIEYTBEJB', NULL, NULL, 'JFP5V6WB4KCCWEHPHBYZQLHLY7KERCU5', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 12:00:23.557543');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('armend.lesi+1@gmail.com', 'ACML3UAQAO2BGBT2IKNOF32R5AX4YKQ6GTHTBYINVKRSR24DHQA3VGVXAQOZMJQZ6SSLSLMOT4WMHAQ=', 'Y3XIFCTE7SNNZSIYEYOOTTF43WCRQ4GX', NULL, NULL, 'TAYJBPWWRBYIMSCDQPVP3I6JZLYHIIR2', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 11:58:43.222224');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('armend.lesi+2@gmail.com', 'AAAR63SNBRAYTOIWTRB3PZNYM52M54H43UH4HTNALNEO7XWCYZGKQFWBWMMJILX6IJC6TZ5BUYPP2VQ=', 'MGRKUZF6WEO4IHEBQISJY3PD4MURKIRY', NULL, NULL, 'JPADP2DDLUFB6TG7ULUHBIKWJRIQTCZU', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 11:59:26.367827');
INSERT INTO public.auth (email, password, refresh_token, email_token, forget_email_token, invite_token, sms, sms_verified, totp, totp_verified, error_count, meta, created_at) VALUES ('armend.lesi+3@gmail.com', 'ABLA4IQBKDIFFHF3QXYNBFGZ5QLWHYBJV3XBJIRVABTX4OCWTARVT3MYOK255TVJIOPY2IJQ4YMD5EY=', '6LCJNYC77G55OVX5ICTJCLUUPJXUSJBO', NULL, NULL, 'IWBVSWM7EKFWZHM7F67D7VUN65GLZQSM', NULL, NULL, NULL, NULL, 0, NULL, '2021-10-30 11:59:59.690664');

-- insert daily_payment
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('e3aaea5d-ed9c-4d42-9132-c09e492b7b7a', 'a457216c-8da3-4b22-8f4d-0fcc8b58328a', 'eth', 76426, 365, '2021-10-30 14:08:12.099251');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('cd4e512a-a84e-48f1-89b6-520b255aeff4', '790f85ab-1fd1-455c-a788-913e9d1b67cb', 'eth', 76439, 730, '2021-10-30 14:09:12.935132');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('4105e1e9-4dce-4af8-ac57-9a1b75a365be', '0e26bf43-b72f-4ac9-9a64-64ec71825aca', 'eth', 76491, 365, '2021-10-30 14:16:08.703374');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('bee34584-4dd7-47d3-95a2-603667f96ace', '0fb5a829-4b4a-4853-b742-679977bd2314', 'eth', 76491, 365, '2021-10-30 14:17:37.707852');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('7fdbfeea-ac4d-4b56-8725-4dfa0383cddf', '0fb5a829-4b4a-4853-b742-679977bd2314', 'xtz', 50813486, 365, '2021-10-30 14:17:37.713275');
INSERT INTO public.daily_payment (id, payment_cycle_id, currency, amount, days_left, last_update) VALUES ('2115e768-6146-4926-befe-36daac097bca', 'e9db8eba-dfd2-4599-b300-365b19812261', 'xtz', 50790468, 365, '2021-10-30 14:23:02.682054');


-- insert invoice
INSERT INTO public.invoice (id, nowpayments_invoice_id, payment_cycle_id, payment_id, price_amount, price_currency, pay_amount, pay_currency, actually_paid, outcome_amount, outcome_currency, payment_status, created_at, last_update) VALUES ('61c43566-86ed-472c-b211-7a1eb7018bf4', 4647051139, 'a457216c-8da3-4b22-8f4d-0fcc8b58328a', 4734113644, 125000000, 'usd', 28895720, 'eth', 29895720, 27895720, 'eth', 'finished', '2021-10-30 12:06:48.236', '2021-10-30 12:08:12');
INSERT INTO public.invoice (id, nowpayments_invoice_id, payment_cycle_id, payment_id, price_amount, price_currency, pay_amount, pay_currency, actually_paid, outcome_amount, outcome_currency, payment_status, created_at, last_update) VALUES ('bd46c888-6f10-4517-b223-a27f89a43d3d', 5372407758, '790f85ab-1fd1-455c-a788-913e9d1b67cb', 5785558822, 125000000, 'usd', 28905100, 'eth', 29905100, 27905100, 'eth', 'finished', '2021-10-30 12:08:18.139', '2021-10-30 12:09:12');
INSERT INTO public.invoice (id, nowpayments_invoice_id, payment_cycle_id, payment_id, price_amount, price_currency, pay_amount, pay_currency, actually_paid, outcome_amount, outcome_currency, payment_status, created_at, last_update) VALUES ('52f431f2-6b0f-4d2c-a54d-3b9875e7ed03', 4323236587, '0e26bf43-b72f-4ac9-9a64-64ec71825aca', 5127184013, 125000000, 'usd', 28919460, 'eth', 29919460, 27919460, 'eth', 'finished', '2021-10-30 12:15:16.662', '2021-10-30 12:16:08');
INSERT INTO public.invoice (id, nowpayments_invoice_id, payment_cycle_id, payment_id, price_amount, price_currency, pay_amount, pay_currency, actually_paid, outcome_amount, outcome_currency, payment_status, created_at, last_update) VALUES ('848d0be7-460b-467b-aaf5-40d220bc57b0', 5448690410, '0fb5a829-4b4a-4853-b742-679977bd2314', 6134730376, 125000000, 'usd', 19546922600, 'xtz', 20546922600, 18546922600, 'xtz', 'finished', '2021-10-30 12:16:16.184', '2021-10-30 12:17:37');
INSERT INTO public.invoice (id, nowpayments_invoice_id, payment_cycle_id, payment_id, price_amount, price_currency, pay_amount, pay_currency, actually_paid, outcome_amount, outcome_currency, payment_status, created_at, last_update) VALUES ('7b626ccc-7f0c-45e3-911d-549eb5bb94c9', 6173885213, 'e9db8eba-dfd2-4599-b300-365b19812261', 6284361591, 125000000, 'usd', 19538520900, 'xtz', 20538520900, 18538520900, 'xtz', 'finished', '2021-10-30 12:22:08.422', '2021-10-30 12:23:02');


-- insert payment_cycle
INSERT INTO public.payment_cycle (id, user_id, days_left, seats, freq, created_at) VALUES ('a457216c-8da3-4b22-8f4d-0fcc8b58328a', '67d62066-b966-425d-8060-2a58a17b48c9', 365,  1, 365, '2021-10-30 12:06:48.13377');
INSERT INTO public.payment_cycle (id, user_id, days_left, seats, freq, created_at) VALUES ('790f85ab-1fd1-455c-a788-913e9d1b67cb', '67d62066-b966-425d-8060-2a58a17b48c9', 730,  1, 365, '2021-10-30 12:08:18.145275');
INSERT INTO public.payment_cycle (id, user_id, days_left, seats, freq, created_at) VALUES ('0e26bf43-b72f-4ac9-9a64-64ec71825aca', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', 365,  1, 365, '2021-10-30 12:15:16.547102');
INSERT INTO public.payment_cycle (id, user_id, days_left, seats, freq, created_at) VALUES ('0fb5a829-4b4a-4853-b742-679977bd2314', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', 730,  1, 365, '2021-10-30 12:16:16.194388');
INSERT INTO public.payment_cycle (id, user_id, days_left, seats, freq, created_at) VALUES ('e9db8eba-dfd2-4599-b300-365b19812261', 'd994260b-125e-441a-926a-bd0498bfb902', 365,  1, 365, '2021-10-30 12:22:08.291194');

-- insert user_balances
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('f3c9f27e-278e-4487-a512-7e0c8e8b45bf', 'a457216c-8da3-4b22-8f4d-0fcc8b58328a', '67d62066-b966-425d-8060-2a58a17b48c9', NULL, 27895720, 'PAYMENT', 'eth', '1970-01-01', '2021-10-30 12:08:12.091596');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('87882c41-f0fb-4a97-b86f-e8a071f5e8cf', 'a457216c-8da3-4b22-8f4d-0fcc8b58328a', '67d62066-b966-425d-8060-2a58a17b48c9', NULL, -27895720, 'CLOSE_CYCLE', 'eth', '1970-01-01', '2021-10-30 12:09:12.920562');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('a7fde994-1c4d-4b74-aaf6-8145568ea5d7', '790f85ab-1fd1-455c-a788-913e9d1b67cb', '67d62066-b966-425d-8060-2a58a17b48c9', NULL, 27895720, 'CARRY_OVER', 'eth', '1970-01-01', '2021-10-30 12:09:12.920562');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('28e3acf8-e9ac-47b0-b3db-fada4af47260', '790f85ab-1fd1-455c-a788-913e9d1b67cb', '67d62066-b966-425d-8060-2a58a17b48c9', NULL, 27905100, 'PAYMENT', 'eth', '1970-01-01', '2021-10-30 12:09:12.920562');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('6f937fd9-4886-434d-af08-186f169f15b9', '0e26bf43-b72f-4ac9-9a64-64ec71825aca', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', NULL, 27919460, 'PAYMENT', 'eth', '1970-01-01', '2021-10-30 12:16:08.700088');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('f5f2518b-2624-4df4-a4c2-012f546c9c77', '0e26bf43-b72f-4ac9-9a64-64ec71825aca', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', NULL, -27919460, 'CLOSE_CYCLE', 'eth', '1970-01-01', '2021-10-30 12:17:37.697472');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('40312518-d187-427d-9c8f-b266822bfa08', '0fb5a829-4b4a-4853-b742-679977bd2314', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', NULL, 27919460, 'CARRY_OVER', 'eth', '1970-01-01', '2021-10-30 12:17:37.697472');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('621a41b0-720e-4207-b778-ea8b7c8bf011', '0fb5a829-4b4a-4853-b742-679977bd2314', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', NULL, 18546922600, 'PAYMENT', 'xtz', '1970-01-01', '2021-10-30 12:17:37.703987');
INSERT INTO public.user_balances (id, payment_cycle_id, user_id, from_user_id, balance, balance_type, currency, day, created_at) VALUES ('3bbaaedf-68a9-4127-a7c6-e504469cd55c', 'e9db8eba-dfd2-4599-b300-365b19812261', 'd994260b-125e-441a-926a-bd0498bfb902', NULL, 18538520900, 'PAYMENT', 'xtz', '1970-01-01', '2021-10-30 12:23:02.672233');


-- insert users
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('7a760bac-5d84-498b-9757-5c913d35c605', NULL, NULL, NULL, NULL, NULL, NULL, 'armend.lesi+4@gmail.com', NULL, NULL, NULL, '6IREUFR7ST3UENSNKOBOWMSIXMMXU===', 'USR', '2021-10-30 12:00:53.293937');
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('67d62066-b966-425d-8060-2a58a17b48c9', NULL, NULL, NULL, NULL, NULL, null, 'armend.lesi+1@gmail.com', NULL, NULL, NULL, 'BW56QCVUHPPTCM5DMAT7K774X4PDA===', 'USR', '2021-10-30 11:59:02.935403');
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('23fbded4-6d6b-446e-8b33-9cefcacc9a01', NULL, NULL, NULL, NULL, NULL, null, 'armend.lesi+2@gmail.com', NULL, NULL, NULL, 'CD2AZO2S7Y73MR76Y6PLKPIYAF3QU===', 'USR', '2021-10-30 11:59:42.816161');
INSERT INTO public.users (id, sponsor_id, invited_email, stripe_id, stripe_payment_method, stripe_last4, payment_cycle_id, email, name, image, payout_eth, token, role, created_at) VALUES ('d994260b-125e-441a-926a-bd0498bfb902', NULL, NULL, NULL, NULL, NULL, null, 'armend.lesi+3@gmail.com', NULL, NULL, NULL, 'YQQYQODRNSLDMONVEDLMUVE34T4AC===', 'USR', '2021-10-30 12:00:12.234188');

update users set payment_cycle_id = '790f85ab-1fd1-455c-a788-913e9d1b67cb' where id = '67d62066-b966-425d-8060-2a58a17b48c9';
update users set payment_cycle_id = '0fb5a829-4b4a-4853-b742-679977bd2314' where id = '23fbded4-6d6b-446e-8b33-9cefcacc9a01';
update users set payment_cycle_id = 'e9db8eba-dfd2-4599-b300-365b19812261' where id = 'd994260b-125e-441a-926a-bd0498bfb902';


-- insert analysis_request
INSERT INTO public.analysis_request (id, repo_id, date_from, date_to, branch, created_at) VALUES ('220b2ad9-8173-4868-85e1-a78054862c7a', 'd081217c-9833-4372-bdf3-a5fc863c1ed3', '2021-07-30', '2021-10-30', 'main', '2021-10-30 12:26:12.330683');
INSERT INTO public.analysis_request (id, repo_id, date_from, date_to, branch, created_at) VALUES ('5a1385a8-9770-475d-b55c-70de4f7ba9ff', '6ed6a9dc-0434-4408-8412-d06baa897c25', '2021-07-30', '2021-10-30', 'master', '2021-10-30 12:26:33.814982');
INSERT INTO public.analysis_request (id, repo_id, date_from, date_to, branch, created_at) VALUES ('585346be-3f90-418e-afce-7a669055c22e', 'bc4c1eb3-3e6c-4d4a-baaa-0cc521a46730', '2021-07-30', '2021-10-30', 'develop', '2021-10-30 12:27:00.661169');
INSERT INTO public.analysis_request (id, repo_id, date_from, date_to, branch, created_at) VALUES ('12f61cd1-42a6-4bb4-b238-5eb58d95a701', '8ddd5e4e-38db-40fb-a93d-5d64a5b13a39', '2021-07-30', '2021-10-30', 'master', '2021-10-30 12:28:35.397585');
INSERT INTO public.analysis_request (id, repo_id, date_from, date_to, branch, created_at) VALUES ('a7e88ad7-9243-4209-a297-22282b4b51b7', 'b0f2d92c-3891-41ad-bca2-48deb5ed0011', '2021-07-30', '2021-10-30', 'dev', '2021-10-30 12:38:30.370413');
INSERT INTO public.analysis_request (id, repo_id, date_from, date_to, branch, created_at) VALUES ('bf5e056a-c880-4ff2-a5bf-c4460a54923c', 'd0a4aeb7-9430-411b-b944-65afdee0b0f4', '2021-07-30', '2021-10-30', 'main', '2021-10-30 12:38:52.063766');

-- insert git email
INSERT INTO public.git_email (id, user_id, email, token, confirmed_at, created_at) VALUES ('08d3a39c-5b41-44bb-b7f6-0995a0d51c67', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', 'armend.lesi+2@gmail.com', 'QMVZMOYGBJPREO2CFCQ6JAFUXQ======', NULL, '2021-10-30 12:48:33.806974');
INSERT INTO public.git_email (id, user_id, email, token, confirmed_at, created_at) VALUES ('6a27baf5-3ed4-41f7-8317-94d844ac75b2', '67d62066-b966-425d-8060-2a58a17b48c9', 'armend.lesi+1@gmail.com', 'C64IVABILT4SRAROBBWDNMVY2E======', NULL, '2021-10-30 12:48:59.413386');
INSERT INTO public.git_email (id, user_id, email, token, confirmed_at, created_at) VALUES ('88ad9ea8-61b9-46b3-8d81-abaa03bc81c1', 'd994260b-125e-441a-926a-bd0498bfb902', 'armend.lesi+3@gmail.com', 'MUWBX6P4OSGG47DCN4KDMOKP24======', NULL, '2021-10-30 12:49:24.488047');
INSERT INTO public.git_email (id, user_id, email, token, confirmed_at, created_at) VALUES ('d3bcf76c-35d6-462e-9c2d-846600710b12', '7a760bac-5d84-498b-9757-5c913d35c605', 'armend.lesi+4@gmail.com', 'MZKRSYENLDMNH6SOCFTFLEVXAY======', NULL, '2021-10-30 12:49:38.526135');

-- insert repo
INSERT INTO public.repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) VALUES ('bc4c1eb3-3e6c-4d4a-baaa-0cc521a46730', 1424470, 'https://github.com/moment/moment', 'https://github.com/moment/moment.git', 'develop', 'moment/moment', 'Parse, validate, manipulate, and display dates in javascript.', '\x0eff81040102ff8200010c010c000004ff820000', 1, 'github', '2021-10-30 12:27:00.652905');
INSERT INTO public.repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) VALUES ('b0f2d92c-3891-41ad-bca2-48deb5ed0011', 11730342, 'https://github.com/vuejs/vue', 'https://github.com/vuejs/vue.git', 'dev', 'vuejs/vue', '🖖 Vue.js is a progressive, incrementally-adoptable JavaScript framework for building UI on the web.', '\x0eff81040102ff8200010c010c000004ff820000', 1, 'github', '2021-10-30 12:38:30.358777');
INSERT INTO public.repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) VALUES ('d0a4aeb7-9430-411b-b944-65afdee0b0f4', 411604532, 'https://github.com/faragos/cloud10', 'https://github.com/faragos/cloud10.git', 'main', 'faragos/cloud10', 'The shop which is one above Cloud 9', '\x0eff81040102ff8200010c010c000004ff820000', 1, 'github', '2021-10-30 12:38:52.050538');
INSERT INTO public.repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) VALUES ('d081217c-9833-4372-bdf3-a5fc863c1ed3', 10270250, 'https://github.com/facebook/react', 'https://github.com/facebook/react.git', 'main', 'facebook/react', 'A declarative, efficient, and flexible JavaScript library for building user interfaces.', '\x0eff81040102ff8200010c010c000004ff820000', 1, 'github', '2021-10-30 12:26:12.322133');
INSERT INTO public.repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) VALUES ('6ed6a9dc-0434-4408-8412-d06baa897c25', 24195339, 'https://github.com/angular/angular', 'https://github.com/angular/angular.git', 'master', 'angular/angular', 'The modern web developer’s platform', '\x0eff81040102ff8200010c010c000004ff820000', 1, 'github', '2021-10-30 12:26:33.80109');
INSERT INTO public.repo (id, orig_id, url, git_url, branch, name, description, tags, score, source, created_at) VALUES ('8ddd5e4e-38db-40fb-a93d-5d64a5b13a39', 41187219, 'https://github.com/jquense/react-big-calendar', 'https://github.com/jquense/react-big-calendar.git', 'master', 'jquense/react-big-calendar', 'gcal/outlook like calendar component', '\x0eff81040102ff8200010c010c000004ff820000', 1, 'github', '2021-10-30 12:28:35.387389');

-- insert sponsor_event
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('ee49595f-b661-4459-ba91-005efcbde80b', 'd081217c-9833-4372-bdf3-a5fc863c1ed3', '67d62066-b966-425d-8060-2a58a17b48c9', '2021-10-30 12:26:12.325272', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('f6d4427c-1601-49d4-bb81-57df8dcbc192', '6ed6a9dc-0434-4408-8412-d06baa897c25', '67d62066-b966-425d-8060-2a58a17b48c9', '2021-10-30 12:26:33.807342', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('ad688eb9-24f0-4322-bf76-ece9756659ef', 'bc4c1eb3-3e6c-4d4a-baaa-0cc521a46730', '67d62066-b966-425d-8060-2a58a17b48c9', '2021-10-30 12:27:00.655763', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('1e2f9b0c-261b-4deb-92b5-f3553dd83cc0', '8ddd5e4e-38db-40fb-a93d-5d64a5b13a39', '67d62066-b966-425d-8060-2a58a17b48c9', '2021-10-30 12:28:35.391506', '2021-10-30 12:28:45.523453');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('2dbab59c-5d89-4e9a-9a1e-0f5044aa6f29', '8ddd5e4e-38db-40fb-a93d-5d64a5b13a39', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', '2021-10-30 12:29:11.89094', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('ff669803-b36a-4d3b-a2d8-3d026c0136d5', 'b0f2d92c-3891-41ad-bca2-48deb5ed0011', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', '2021-10-30 12:38:30.363186', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('a4cb7038-1055-4f3d-8d61-3653fd871569', 'd0a4aeb7-9430-411b-b944-65afdee0b0f4', 'd994260b-125e-441a-926a-bd0498bfb902', '2021-10-30 12:38:52.058235', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('f54173bb-d683-4cf5-a026-0c49ad95acc3', 'd081217c-9833-4372-bdf3-a5fc863c1ed3', 'd994260b-125e-441a-926a-bd0498bfb902', '2021-10-30 12:39:19.154767', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('99812b4c-d7c5-4ba2-82c9-12d0fd2b4da3', '6ed6a9dc-0434-4408-8412-d06baa897c25', 'd994260b-125e-441a-926a-bd0498bfb902', '2021-10-30 12:39:24.896692', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('5c1cf33c-feca-4c17-9699-056c626de97b', 'd081217c-9833-4372-bdf3-a5fc863c1ed3', '7a760bac-5d84-498b-9757-5c913d35c605', '2021-10-30 12:39:45.533555', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('cdd15b00-72cf-485d-983e-b86b98980a03', '6ed6a9dc-0434-4408-8412-d06baa897c25', '7a760bac-5d84-498b-9757-5c913d35c605', '2021-10-30 12:39:51.219999', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('b0188c0e-dfd7-4bcb-9a40-1dd80bcce133', '6ed6a9dc-0434-4408-8412-d06baa897c25', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', '2021-10-30 12:40:05.266736', '9999-01-01 00:00:00');
INSERT INTO public.sponsor_event (id, repo_id, user_id, sponsor_at, unsponsor_at) VALUES ('775b13c3-4444-4f0b-b523-c6a622324aba', '8ddd5e4e-38db-40fb-a93d-5d64a5b13a39', '67d62066-b966-425d-8060-2a58a17b48c9', '2021-10-30 12:40:30.542324', '9999-01-01 00:00:00');


-- insert user_emails_sent
INSERT INTO public.user_emails_sent (id, user_id, email_type, created_at) VALUES ('e3bfc06e-a520-4112-aa3c-207475bb1467', '67d62066-b966-425d-8060-2a58a17b48c9', 'gitemail-armend.lesi%40gmail.com', '2021-10-30 12:47:55.139805');
INSERT INTO public.user_emails_sent (id, user_id, email_type, created_at) VALUES ('5f32fa8e-0747-4ab9-ba42-3a767a539e58', '23fbded4-6d6b-446e-8b33-9cefcacc9a01', 'gitemail-armend.lesi%2B2%40gmail.com', '2021-10-30 12:48:33.81012');
INSERT INTO public.user_emails_sent (id, user_id, email_type, created_at) VALUES ('2a6c4e5c-04a1-4122-939e-d555e1056414', '67d62066-b966-425d-8060-2a58a17b48c9', 'gitemail-armend.lesi%2B1%40gmail.com', '2021-10-30 12:48:59.415653');
INSERT INTO public.user_emails_sent (id, user_id, email_type, created_at) VALUES ('e6df6443-c326-4197-b451-2c4150acaad3', 'd994260b-125e-441a-926a-bd0498bfb902', 'gitemail-armend.lesi%2B3%40gmail.com', '2021-10-30 12:49:24.489991');
INSERT INTO public.user_emails_sent (id, user_id, email_type, created_at) VALUES ('c534abff-2cb9-4a06-adeb-f6615616a59b', '7a760bac-5d84-498b-9757-5c913d35c605', 'gitemail-armend.lesi%2B4%40gmail.com', '2021-10-30 12:49:38.529796');



-- tests
CREATE OR REPLACE FUNCTION updateDailyUserBalance(yesterdayStart DATE, now TIMESTAMP with time zone) RETURNS SETOF record AS
$$
DECLARE
r record;
	_id uuid;
BEGIN
FOR r IN
SELECT
    u.payment_cycle_id,
    u.id as user_id,
    -dp.amount as balance,
    'DAY' as balance_type,
    dp.currency,
    yesterdayStart as day,
			now as created_at
FROM daily_payment dp
    INNER JOIN users u on u.payment_cycle_id = dp.payment_cycle_id
    INNER JOIN daily_repo_hours drh ON u.id = drh.user_id
WHERE drh.day = yesterdayStart
order by dp.payment_cycle_id, days_left asc
    LOOP
    if _id = r.payment_cycle_id then
    continue;
end if;

_id = r.payment_cycle_id;
        RETURN NEXT r; -- return current row of SELECT
END LOOP;
    RETURN;
END;
$$
LANGUAGE plpgsql;
--INSERT INTO user_balances (payment_cycle_id, user_id, balance, balance_type, currency, day, created_at)
select * from test(to_date('2021-10-29', 'YYYY-MM-DD'), now()) f(payment_cycle_id uuid, user_id uuid, balance bigint, balance_type text, currency VARCHAR(16), day date, created_at TIMESTAMP with time zone)

-- select * from daily_repo_hours
--
-- select * from user_balances order by created_at desc

select * from user_balances order by created_at desc

delete from user_balances where id = 'cd111bc8-860a-473c-a1be-acf2705fc5b5'



select
    s.repo_id,
    min(q.payPerHour),
    q.currency,
    SUM((EXTRACT(epoch from age(LEAST('2021-10-30 23:59:59.325272', s.unsponsor_at), GREATEST('2021-10-30 00:00:00.00000', s.sponsor_at)))/3600)::bigint),
    SUM(((EXTRACT(epoch from age(LEAST('2021-10-30 23:59:59.325272', s.unsponsor_at), GREATEST('2021-10-30 00:00:00.00000', s.sponsor_at)))/3600)::bigint * q.payPerHour)
		* (EXTRACT(epoch from age(LEAST('2021-10-30 23:59:59.325272', s.unsponsor_at), GREATEST('2021-10-30 00:00:00.00000', s.sponsor_at)))/3600)::bigint / drh.repo_hours)
from (
         select ub.user_id,-(min(ub.balance) / 24) as payPerHour, min(ub.currency) as currency from user_balances ub
                                                                                                        join sponsor_event s on s.user_id = ub.user_id
         where ub.day = '2021-10-29 00:26:12.325272'
           AND NOT((s.sponsor_at<'2021-10-30 00:26:12.325272' AND s.unsponsor_at<'2021-10-30 00:26:12.325272') OR (s.sponsor_at>='2021-10-30 23:26:12.325272' AND s.unsponsor_at>='2021-10-30 23:26:12.325272'))
           AND ub.balance_type = 'DAY'
         group by ub.user_id
     ) as q
         join sponsor_event s on s.user_id = q.user_id
         INNER JOIN users u ON u.id = s.user_id
         INNER JOIN daily_repo_hours drh ON u.id = drh.user_id
group by s.repo_id, q.currency
