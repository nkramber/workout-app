# Workout App - exercise safety research

Status: research input for the roadmap. This document holds evidence and synthesis. It holds no owner decision. The owner decisions live in `docs/decisions.md`.

Date of this document: 2026-09-28. Access date of every source: 2026-09-27.

## 1. Purpose

This document gives the deterministic policy of D-23 a cited base. Luna proposes plans and revisions (D-22). The policy checks every set, load, and change before the owner sees it (D-23). Each policy rule can name the `EV-<n>` ids that support it, and each assumption stays visible as an assumption.

The scope follows the owner decisions. Workout App serves one user, the owner (D-67). The user is an adult (D-26) with intermediate or advanced experience (D-30). The core user is a 180 lb, 32-year-old man who used machines before but did not train for several years (D-31). The engine treats him as intermediate from the first day, with loads from his own estimates (D-32). Equipment is fixed-path resistance machines and cardio machines only (D-45).

This document gives fitness guidance research only (D-36). It is not medical, legal, or regulatory advice.

## 2. Method

- Evidence standard: current authoritative guidance plus peer-reviewed systematic evidence, with a citation and a date for each item (D-38).
- Peer-reviewed papers: the researcher took records and abstracts from the Europe PMC REST API. For open-access papers, the researcher read the full text through Europe PMC.
- PubMed and PMC HTML pages returned reCAPTCHA pages to the fetch tool. The Europe PMC record stands in for them.
- Guidance and law: the researcher fetched the primary PDFs, Federal Register text, and statute pages, and extracted their text.
- Archive snapshots: hhs.gov (HTTP 403), acog.org (HTTP 402), and the retired EIM form URL blocked direct fetches. For these, the researcher read an Internet Archive snapshot of the same official URL. The register marks each case with "archive snapshot".
- Search snippets, blogs, and law-firm summaries do not count as evidence. When only secondary coverage exists, the item has the class unresolved.
- The raw research targeted novice consumers on a public app. This document applies it again to the single-user, intermediate, machine-only scope.

Classes in the register:

| Class | Meaning |
|---|---|
| evidence | Primary data, or a systematic review or meta-analysis of data. "evidence (legal text)" marks binding statute or regulation text. |
| recommendation | A guideline, position stand, consensus statement, or non-binding regulatory guidance. |
| assumption | A design choice or extrapolation that the sources do not directly support. |
| unresolved | Not verified, or the verified content is incomplete. |

## 3. Limits

- Most trials are small (about 10 to 40 people per arm) and short (6 to 12 weeks).
- Participants are mostly young, male, and supervised (EV-19: 79% male, mean age about 25).
- No study tests an unsupervised program that an LLM adapts on a phone.
- The ACSM 2026 authors report poor adverse-event records in this field (EV-1).
- Many numbers in the synthesis are assumptions. The text marks each one.
- Group-level evidence does not predict the response of one person. The core user can respond above or below the group mean.
- The researcher did not read some full texts and some tables. Section 7 lists them.

## 4. Evidence register

| Id | Source (title, authors/publisher, year, linked URL/DOI) | Supports | Limits | Class |
|---|---|---|---|---|
| EV-1 | "ACSM Position Stand. Resistance Training Prescription for Muscle Function, Hypertrophy, and Physical Performance in Healthy Adults: An Overview of Reviews." Currier BS, D'Souza AC, Fiatarone Singh MA, et al., Phillips SM (senior author). ACSM, Med Sci Sports Exerc 2026;58(4):851-872. [PMC12965823](https://pmc.ncbi.nlm.nih.gov/articles/PMC12965823), [DOI 10.1249/MSS.0000000000003897](https://doi.org/10.1249/MSS.0000000000003897), [ACSM release](https://acsm.org/resistance-training-guidelines-update-2026/) | Overview of 137 reviews, more than 30,000 people, searches to Oct 2024. Any RT beats none. Adherence and individual fit matter more than exact prescription. Strength: 2+ days/week, heavier loads, 2-3 sets, full ROM. Hypertrophy: dose-response at 10+ sets/muscle/week. Failure not required, "near-failure" or 2-3 RIR is sufficient effort. Exact RIR targets cannot be quantified. Machines versus free weights, rest, periodization: no consistent effect. Four regions (upper/lower x push/pull) cover the major groups. RT is safe for healthy adults of all ages, with no rise in serious adverse events in more than 38,000 people. Flags form and vascular risk of lifting to fatigue in some groups. | Group-level. Many inexperienced trainees. Healthy adults only. Possible overlap across reviews. No muscle-specific comparisons. | recommendation |
| EV-2 | "Progression models in resistance training for healthy adults." ACSM position stand, Med Sci Sports Exerc 2009;41:687-708. [DOI 10.1249/MSS.0b013e3181915670](https://doi.org/10.1249/MSS.0b013e3181915670) | Increase load 2-10% when the person exceeds the target reps by 1-2. Large before small, multi-joint before single-joint. Historical basis of double progression. | Superseded by EV-1. Largely expert-driven. Use only for progression-step heuristics. | recommendation |
| EV-3 | Physical Activity Guidelines for Americans, 2nd edition. US HHS, 2018. [PDF](https://odphp.health.gov/sites/default/files/2019-09/Physical_Activity_Guidelines_2nd_edition.pdf), [current guidelines page](https://odphp.health.gov/our-work/nutrition-physical-activity/physical-activity-guidelines/current-guidelines). Summary: Piercy KL et al., JAMA 2018, [DOI 10.1001/jama.2018.14854](https://doi.org/10.1001/jama.2018.14854) | Muscle strengthening, moderate or greater effort, all major groups, 2+ days/week. One set of 8-12 reps works, 2-3 sets can work better. Inactive people: "start low and go slow", increase gradually. Chronic conditions: be under the care of a provider. Pregnancy: provider care. | Health outcomes, not strength optimization. Dated 2018, no 3rd edition found. | recommendation |
| EV-4 | "Adult Activity: An Overview." CDC, last reviewed 20 Dec 2023. [Page](https://www.cdc.gov/physical-activity-basics/guidelines/adults.html) | Muscle strengthening 2+ days/week across legs, hips, back, abdomen, chest, shoulders, arms. | Consumer restatement of EV-3. No dose detail. | recommendation |
| EV-5 | "World Health Organization 2020 guidelines on physical activity and sedentary behaviour." Bull FC et al., Br J Sports Med 2020;54:1451-1462. [DOI 10.1136/bjsports-2020-102955](https://doi.org/10.1136/bjsports-2020-102955), [WHO publication](https://www.who.int/publications/i/item/9789240015128) | Muscle strengthening 2+ days/week for adults, including chronic conditions and disability. Medical clearance generally unnecessary for light or moderate activity without contraindications. New symptoms: consult a provider. Older adults: multicomponent, balance, 3+ days/week. Pregnancy: no supine work after the first trimester, know the danger signs. | Health-outcome focus. No RT variable detail. | recommendation |
| EV-6 | ACSM's Guidelines for Exercise Testing and Prescription, 12th edition (GETP12). ACSM, 2025. [Announcement](https://acsm.org/certification-exam-2025-getp12/) | GETP12 exists and has been the exam basis since 10 Jul 2025. No noteworthy practice change, so the EV-7 screening model is likely still current. | The book was not accessed. Termination criteria and symptom tables not verified. | recommendation, content unresolved |
| EV-7 | "Updating ACSM's Recommendations for Exercise Preparticipation Health Screening." Riebe D, Franklin BA, Thompson PD, et al., Med Sci Sports Exerc 2015;47:2473-2479. [DOI 10.1249/MSS.0000000000000664](https://doi.org/10.1249/MSS.0000000000000664) | Screening rests on current activity, signs or symptoms or known cardiovascular, metabolic, or renal disease, and desired intensity. Cardiac events "are often preceded by warning signs/symptoms". | Abstract only (full text paywalled). Cardiovascular risk only, not musculoskeletal. | recommendation |
| EV-8 | "Exercise Preparticipation Health Screening Questionnaire for Exercise Professionals." ACSM/Exercise is Medicine, from Magal M, Riebe D, ACSM Health Fit J 2016;20(3):22-27. [Original URL, archive snapshot](https://www.exerciseismedicine.org/assets/page_documents/EIM%20exercise%20preparticipation%20screening.pdf) | Step 1 symptoms: chest discomfort with exertion, unreasonable breathlessness, dizziness, fainting, blackouts, ankle swelling, forceful or irregular heart rate, calf burning on short walks, known heart murmur. Any symptom: "STOP", seek medical clearance before exercise. Step 3 conditions plus inactivity: clearance recommended. | Built for exercise professionals, not self-report. Cardiovascular, metabolic, renal only. Copyrighted. Archive snapshot. | recommendation |
| EV-9 | "The Physical Activity Readiness Questionnaire for Everyone (PAR-Q+)" 2025. PAR-Q+ Collaboration, version dated 01-11-2024. [Site](https://eparmedx.com/), [PDF](https://eparmedx.com/wp-content/uploads/2025/01/PARQPlus2025Fillable.pdf) | Seven questions: heart condition or high BP, chest pain, dizziness or loss of consciousness, other chronic condition, chronic medication, bone or joint problem, doctor advice for supervised activity only. Delay for temporary illness or pregnancy. Age 45+ and not used to vigorous effort: consult a professional first. Clearance valid 12 months. | Self-report. Licensing for app use unresolved. ePARmed-X+ logic not public. No 2026 edition found. | recommendation |
| EV-10 | "PAR-Q+ and ePARmed-X+: new risk stratification and physical activity clearance strategy for physicians and patients alike." Bredin SSD et al., Can Fam Physician 2013;59:273-277. [PMC3596208](https://europepmc.org/article/PMC/PMC3596208) | Background on the PAR-Q+ risk pathway. | Record only. Validation studies not reviewed. | recommendation |
| EV-11 | "Physical Activity and Exercise During Pregnancy and the Postpartum Period." ACOG Committee Opinion 804, 2020, reaffirmed 2023. Obstet Gynecol 2020;135:e178-e188. [DOI 10.1097/AOG.0000000000003772](https://doi.org/10.1097/AOG.0000000000003772), [ACOG page, archive snapshot](https://www.acog.org/clinical/clinical-guidance/committee-opinion/articles/2020/04/physical-activity-and-exercise-during-pregnancy-and-the-postpartum-period) | Clinical evaluation before a program. Resistance exercise safe in uncomplicated pregnancy. Warning signs to stop: bleeding, abdominal pain, contractions, fluid leakage, dyspnea before exertion, dizziness, headache, chest pain, weakness, calf pain or swelling. | Contraindication tables are images, not transcribed. Archive snapshot. | recommendation |
| EV-12 | "2019 Canadian guideline for physical activity throughout pregnancy." Mottola MF et al., Br J Sports Med 2018;52:1339-1346. [DOI 10.1136/bjsports-2018-100056](https://doi.org/10.1136/bjsports-2018-100056) | A second GRADE-based prenatal guideline exists. | Abstract only. Lists not extracted. | recommendation, content unresolved |
| EV-13 | "Resistance Training for Older Adults: Position Statement From the National Strength and Conditioning Association." Fragala MS et al., J Strength Cond Res 2019;33:2019-2052. [DOI 10.1519/JSC.0000000000003230](https://doi.org/10.1519/JSC.0000000000003230) | RT is a powerful intervention against age-related loss of strength and muscle. | Abstract only. Dosing unresolved. | recommendation |
| EV-14 | "Progressive resistance strength training for improving physical function in older adults." Liu CJ, Latham NK, Cochrane Database Syst Rev 2009. [DOI 10.1002/14651858.CD002759.pub2](https://doi.org/10.1002/14651858.CD002759.pub2) | 121 trials, n=6,700. Progressive RT at 2-3 times/week improved physical ability, gait speed, chair rise. | Dated. Heterogeneous trials. Poor adverse-event reports. | evidence |
| EV-15 | Youth position statements. Lloyd RS et al., "Position statement on youth resistance training: the 2014 International Consensus", Br J Sports Med 2014;48:498-505, [DOI 10.1136/bjsports-2013-092952](https://doi.org/10.1136/bjsports-2013-092952). Faigenbaum AD et al., NSCA youth position paper, J Strength Cond Res 2009;23:S60-79, [DOI 10.1519/JSC.0b013e31819df407](https://doi.org/10.1519/JSC.0b013e31819df407) | Youth RT is endorsed when "appropriately prescribed and supervised". Supports the adult-only scope. | Scope exclusion use only. | recommendation |
| EV-16 | "Effects of Resistance Training Frequency on Measures of Muscle Hypertrophy: A Systematic Review and Meta-Analysis." Schoenfeld BJ, Ogborn D, Krieger JW, Sports Med 2016;46:1689-1697. [DOI 10.1007/s40279-016-0543-8](https://doi.org/10.1007/s40279-016-0543-8) | Volume-equated, 2 times/week per muscle beat 1 time/week for hypertrophy. | 10 studies. 3 versus 2 times undetermined. | evidence |
| EV-17 | "How many times per week should a muscle be trained to maximize muscle hypertrophy?" Schoenfeld BJ, Grgic J, Krieger J, J Sports Sci 2019;37:1286-1295. [DOI 10.1080/02640414.2018.1555906](https://doi.org/10.1080/02640414.2018.1555906) | 25 studies. Volume-equated frequency has no meaningful effect on hypertrophy. Frequency can follow preference. | Heterogeneous, short-term. | evidence |
| EV-18 | "Dose-response relationship between weekly resistance training volume and increases in muscle mass." Schoenfeld BJ, Ogborn D, Krieger JW, J Sports Sci 2017;35:1073-1082. [DOI 10.1080/02640414.2016.1210197](https://doi.org/10.1080/02640414.2016.1210197) | Each added weekly set gave about 0.37% more hypertrophy. Trend across under 5, 5-9, and 10+ sets/muscle/week. | 15 studies. | evidence |
| EV-19 | "The Resistance Training Dose Response: Meta-Regressions Exploring the Effects of Weekly Volume and Frequency on Muscle Hypertrophy and Strength Gains." Pelland JC, Remmert JF, Robinson ZP, Hinson SR, Zourdos MC, Sports Med 2026. [DOI 10.1007/s40279-025-02344-w](https://doi.org/10.1007/s40279-025-02344-w), [preprint](https://sportrxiv.org/index.php/server/preprint/view/460) | 67 studies, 2,058 people. Volume raises hypertrophy and strength with diminishing returns, stronger for strength. Indirect sets count best as 0.5. Frequency: negligible for hypertrophy, positive for strength. | 79% male, mean age about 25. Diminishing-return points not extracted. | evidence |
| EV-20 | "Strength and Hypertrophy Adaptations Between Low- vs. High-Load Resistance Training." Schoenfeld BJ et al., J Strength Cond Res 2017;31:3508-3523. [DOI 10.1519/JSC.0000000000002200](https://doi.org/10.1519/JSC.0000000000002200) | 21 studies. Heavier loads gave more 1RM gain. Hypertrophy similar across loads. | All sets to failure. | evidence |
| EV-21 | "Resistance Training Load Effects on Muscle Hypertrophy and Strength Gain." Lopez P et al., Med Sci Sports Exerc 2021;53:1206-1216. [DOI 10.1249/MSS.0000000000002585](https://doi.org/10.1249/MSS.0000000000002585) | 28 studies. Low, moderate, high loads gave similar hypertrophy. Strength favored moderate and high loads. | Failure-based protocols. Small studies. | evidence |
| EV-22 | "Maximal Number of Repetitions at Percentages of the One Repetition Maximum." Nuzzo JL et al., Sports Med 2024;54:303-321. [DOI 10.1007/s40279-023-01937-7](https://doi.org/10.1007/s40279-023-01937-7) | 269 studies, 7,289 people. Leg press about 9/13/19 reps at 90/80/70% 1RM. Between-person SD about 2.5 reps at 80% and 4.4 at 60%. A fixed %1RM gives very different effort across people. | Mostly bench and leg press. 66% male. | evidence |
| EV-23 | "Effect of Resistance Training to Muscle Failure vs. Non-Failure on Strength, Hypertrophy, and Muscle Architecture." Grgic J et al., J Sport Health Sci 2022;11:202-211. [DOI 10.1016/j.jshs.2021.01.007](https://doi.org/10.1016/j.jshs.2021.01.007) | 15 studies. No difference for strength or hypertrophy. Non-failure favored for strength when volume differs. Trained people: small hypertrophy edge for failure. | Young adults only. | evidence |
| EV-24 | "Influence of Resistance Training Proximity-to-Failure on Skeletal Muscle Hypertrophy." Refalo MC et al., Sports Med 2023;53:649-665. [DOI 10.1007/s40279-022-01784-y](https://doi.org/10.1007/s40279-022-01784-y) | Momentary failure not superior to non-failure. Relationship can be non-linear. | Inconsistent failure definitions. | evidence |
| EV-25 | Refalo MC et al., within-subject RCT, failure versus 1-2 RIR, J Sports Sci 2024;42:85-101. [DOI 10.1080/02640414.2024.2321021](https://doi.org/10.1080/02640414.2024.2321021) | Similar quadriceps hypertrophy over 8 weeks. Failure gave more acute fatigue. | n=18, trained, quadriceps only. | evidence |
| EV-26 | "Exploring the Dose-Response Relationship Between Estimated Resistance Training Proximity to Failure, Strength Gain, and Muscle Hypertrophy." Robinson ZP et al., Sports Med 2024;54:2209-2231. [DOI 10.1007/s40279-024-02069-2](https://doi.org/10.1007/s40279-024-02069-2) | Strength similar across a wide RIR range. Hypertrophy improves as sets end closer to failure. | RIR estimated from study text. Authors call it exploratory. | evidence (low certainty) |
| EV-27 | "Accuracy in Predicting Repetitions to Task Failure in Resistance Exercise." Halperin I et al., Sports Med 2022;52:377-390. [DOI 10.1007/s40279-021-01559-x](https://doi.org/10.1007/s40279-021-01559-x) | 12 studies, n=414. People underpredict reps left by about 1. Accuracy better near failure and in sets of 12 or fewer. Training status no effect. | Very high heterogeneity. | evidence |
| EV-28 | Steele J et al., prediction of repetitions to failure by experience, PeerJ 2017;5:e4105. [DOI 10.7717/peerj.4105](https://doi.org/10.7717/peerj.4105) | n=141. Underprediction, SEM about 2.6-3.4 reps. Slightly better accuracy with experience. "RIR should be used cautiously." | Single study. | evidence |
| EV-29 | "Novel Resistance Training-Specific RPE Scale Measuring Repetitions in Reserve." Zourdos MC et al., J Strength Cond Res 2016;30:267-275. [DOI 10.1519/JSC.0000000000001049](https://doi.org/10.1519/JSC.0000000000001049) | RIR-based RPE scale. Novices underrate maximal effort. | n=29, squat only. | evidence |
| EV-30 | Ormsbee MJ et al., RIR-based RPE in bench press, novice versus experienced, J Strength Cond Res 2019;33:337-345. [DOI 10.1519/JSC.0000000000001901](https://doi.org/10.1519/JSC.0000000000001901) | Experience changes RPE accuracy. | College-aged men only. | evidence |
| EV-31 | Droguett et al., RIR accuracy in back squat, J Hum Kinet 2026;102:145-158. [DOI 10.5114/jhk/205218](https://doi.org/10.5114/jhk/205218) | Experience did not change accuracy. Familiarity with high effort can matter more. | n=16. Conflicts with EV-28 to EV-30. | evidence (small study) |
| EV-32 | "Application of the Repetitions in Reserve-Based Rating of Perceived Exertion Scale for Resistance Training." Helms ER et al., Strength Cond J 2016;38:42-49. [DOI 10.1519/SSC.0000000000000218](https://doi.org/10.1519/SSC.0000000000000218) | Practical use of RIR to adjust load set to set. | Narrative. | recommendation |
| EV-33 | Morán-Navarro R et al., "Time course of recovery following resistance training leading or not to failure", Eur J Appl Physiol 2017;117:2387-2399. [DOI 10.1007/s00421-017-3725-7](https://doi.org/10.1007/s00421-017-3725-7) | Failure slowed neuromuscular and metabolic recovery by 24-48 h at equal volume. | n=10 trained men. Acute. | evidence |
| EV-34 | Huynh A et al., exertional rhabdomyolysis after high-intensity RT, Intern Med J 2016;46:602-608. [DOI 10.1111/imj.13055](https://doi.org/10.1111/imj.13055) | 10 of 12 severe exertional cases linked to high-intensity RT programs. Hazard signal for unaccustomed high effort and volume. | Case series. No denominator. | evidence (hazard signal only) |
| EV-35 | Grgic J et al., rest intervals and hypertrophy, Eur J Sport Sci 2017;17:983-993. [DOI 10.1080/17461391.2017.1340524](https://doi.org/10.1080/17461391.2017.1340524) | Short and long rest both work. Trained people can gain from longer rest. | 6 studies. | evidence |
| EV-36 | Grgic J et al., "Effects of Rest Interval Duration in Resistance Training on Measures of Muscular Strength", Sports Med 2018;48:137-151. [DOI 10.1007/s40279-017-0788-x](https://doi.org/10.1007/s40279-017-0788-x) | 23 studies. Trained people need more than 2 min for maximal strength. 60-120 s suffices for untrained people. | Mixed designs. | evidence |
| EV-37 | Singer A et al., Bayesian meta-analysis of rest and hypertrophy, Front Sports Act Living 2024;6:1429789. [DOI 10.3389/fspor.2024.1429789](https://doi.org/10.3389/fspor.2024.1429789) | Small benefit of rest over 60 s. No difference beyond 90 s. | 9 studies. | evidence |
| EV-38 | Plotkin D et al., "Progressive overload without progressing load?", PeerJ 2022;10:e14142. [DOI 10.7717/peerj.14142](https://doi.org/10.7717/peerj.14142) | Rep progression and load progression gave similar adaptations. Supports rep progression when load steps are coarse. | Trained, lower body, 8 weeks, n=43. | evidence |
| EV-39 | Hickmott LM et al., autoregulation meta-analysis, Sports Med Open 2022;8:9. [DOI 10.1186/s40798-021-00404-9](https://doi.org/10.1186/s40798-021-00404-9) | RPE/RIR autoregulation and percentage-based load gave similar 1RM gains. | Trained participants only. | evidence |
| EV-40 | "Integrating Deloading into Strength and Physique Sports Training Programmes: An International Delphi Consensus Approach." Bell L et al., Sports Med Open 2023;9:87. [DOI 10.1186/s40798-023-00633-0](https://doi.org/10.1186/s40798-023-00633-0) | Deload: reduced stress to manage fatigue. Typical every 4-6 weeks, about 7 days. Reduce volume, keep or reduce intensity. Can be planned or autoregulated. | Expert coaches in strength sports. Opinion-level. | recommendation |
| EV-41 | "A Practical Approach to Deloading: Recommendations and Considerations for Strength and Physique Sports." Bell L, Darragh IAJ, Travis SK, Rogerson D, Nolan D, Strength Cond J 2025, ahead of print. [Author PDF](https://doras.dcu.ie/31501/1/a_practical_approach_to_deloading__recommendations.203%282%29.pdf) | Reactive deloads guided by strength performance and perceived readiness. Methods: fewer sets, higher RIR. | Narrative. Athlete focus. DOI not captured. | recommendation |
| EV-42 | Coleman M et al., 1-week deload RCT, PeerJ 2024;12:e16777. [DOI 10.7717/peerj.16777](https://doi.org/10.7717/peerj.16777) | One week of full cessation: same hypertrophy, slightly worse strength. Full stop is not free. | n=39, trained. | evidence |
| EV-43 | Bosquet L et al., training cessation meta-analysis, Scand J Med Sci Sports 2013;23:e140-9. [DOI 10.1111/sms.12047](https://doi.org/10.1111/sms.12047) | 103 studies. Cessation reduces all muscular performance. Loss grows with duration and age. | Abstract truncated. Duration detail unresolved. | evidence |
| EV-44 | Grgic J, detraining in older adults, Int J Environ Res Public Health 2022;19:14048. [DOI 10.3390/ijerph192114048](https://doi.org/10.3390/ijerph192114048) | Muscle size falls with cessation. Non-significant at 12-24 weeks, significant at 31-52 weeks. | 6 studies. Older adults. | evidence |
| EV-45 | Yang et al., strength maintenance after cessation in middle-aged and older adults, J Aging Phys Act 2022;30:552-566. [DOI 10.1123/japa.2020-0493](https://doi.org/10.1123/japa.2020-0493) | Benefits held after 24+ weeks of training even with an equal break. Shorter programs need shorter breaks. | Middle-aged and older adults. | evidence |
| EV-46 | Ogasawara R et al., periodic versus continuous training, Eur J Appl Physiol 2013;113:975-985. [DOI 10.1007/s00421-012-2511-9](https://doi.org/10.1007/s00421-012-2511-9) | 3-week breaks with 6-week retraining matched continuous training. Retraining recovers gains fast. | n=14 young men, bench only. | evidence |
| EV-47 | Psilander N et al., training, detraining, retraining, J Appl Physiol 2019;126:1636-1645. [DOI 10.1152/japplphysiol.00917.2018](https://doi.org/10.1152/japplphysiol.00917.2018) | After 20 weeks off, muscle size returned to baseline, strength stayed partly elevated. | Small sample. | evidence |
| EV-48 | Reynolds JM, Gordon TJ, Robergs RA, prediction of 1RM from multiple-rep tests, J Strength Cond Res 2006;20:584-592. [DOI 10.1519/R-15304.1](https://doi.org/10.1519/R-15304.1) | 5RM predicts 1RM best. Use no more than 10 reps. Error about 3 kg for chest press, about 16 kg for leg press. | n=70. | evidence |
| EV-49 | Mayhew JL et al., 1RM prediction in women, J Strength Cond Res 2008;22:1570-1577. [DOI 10.1519/JSC.0b013e31817b02ad](https://doi.org/10.1519/JSC.0b013e31817b02ad) | Equations more accurate under 10 reps. Large individual differences. | Women only. | evidence |
| EV-50 | Richens B, Cleather DJ, Biol Sport 2014;31:157-161. [DOI 10.5604/20831862.1099047](https://doi.org/10.5604/20831862.1099047) | At 70% 1RM leg press, endurance athletes did about 40 reps, weightlifters about 18. Tables can mislead. | n=16. | evidence |
| EV-51 | Keogh JWL, Winwood PW, "The Epidemiology of Injuries Across the Weight-Training Sports", Sports Med 2017;47:479-501. [DOI 10.1007/s40279-016-0575-0](https://doi.org/10.1007/s40279-016-0575-0) | 0.24-1 injuries per 1,000 h in bodybuilding. Low against team sports. Shoulder, low back, knee, elbow, wrist most often. | Competitive athletes. Mostly retrospective. | evidence |
| EV-52 | Kerr ZY, Collins CL, Comstock RD, ED weight-training injuries 1990-2007, Am J Sports Med 2010;38:765-771. [DOI 10.1177/0363546509351560](https://doi.org/10.1177/0363546509351560) | Weights that drop on the person: 65.5% of injuries. 90.4% involved free weights. | ED only. No exposure denominator. | evidence |
| EV-53 | "Exercise-Related Acute Cardiovascular Events and Potential Deleterious Adaptations Following Long-Term Exercise Training." Franklin BA et al., AHA scientific statement, Circulation 2020;141:e705-e736. [DOI 10.1161/CIR.0000000000000749](https://doi.org/10.1161/CIR.0000000000000749) | Vigorous activity, mainly in unfit people, can acutely raise the risk of sudden cardiac death and MI in susceptible people. | Abstract only. Mainly aerobic exercise. | recommendation |
| EV-54 | Smith BE et al., painful versus pain-free exercise for chronic musculoskeletal pain, Br J Sports Med 2017;51:1679-1687. [DOI 10.1136/bjsports-2016-097383](https://doi.org/10.1136/bjsports-2016-097383) | Pain during exercise: small short-term benefit, no long-term difference, in clinical groups. | Clinical populations. 7 trials. | evidence (outside the app population) |
| EV-55 | Silbernagel KG et al., pain-monitoring model in Achilles tendinopathy, Am J Sports Med 2007;35:897-906. [DOI 10.1177/0363546506298279](https://doi.org/10.1177/0363546506298279) | Clinician-supervised loading under a pain-monitoring model did no harm against active rest. | Clinical population. Not a wellness default. | evidence (clinical population) |
| EV-56 | "General Wellness: Policy for Low Risk Devices." US FDA, final guidance issued 6 Jan 2026, supersedes 2019. [Guidance page](https://www.fda.gov/regulatory-information/search-fda-guidance-documents/general-wellness-policy-low-risk-devices), [PDF](https://www.fda.gov/media/90652/download) | Physical fitness claims are wellness claims: track exercise, improve fitness, strength, muscle size. Non-wellness examples: "treat muscle atrophy", restore function impaired by disease. Labeling and marketing must not exceed the intended use. Sensor outputs must not guide clinical action. | Non-binding. Intended use comes from all claims. | recommendation |
| EV-57 | "Clinical Decision Support Software." US FDA, guidance issued 29 Jan 2026. [Guidance page](https://www.fda.gov/regulatory-information/search-fda-guidance-documents/clinical-decision-support-software), [PDF](https://www.fda.gov/media/109618/download) | The non-device CDS criteria need a health care professional user. Patient-facing software cannot use that exclusion. | Non-binding. | recommendation |
| EV-58 | MDCG 2019-11 Rev.1, "Guidance on Qualification and Classification of Software in Regulation (EU) 2017/745 - MDR and Regulation (EU) 2017/746 - IVDR." European Commission MDCG, Oct 2019, rev.1 Jun 2025. [PDF](https://health.ec.europa.eu/document/download/b45335c5-1679-4c71-a91c-fc7a4d37f12b_en?filename=mdcg_2019_11_en.pdf) | Wellness or fitness apps do not qualify as medical device software. | Non-binding. Turns on stated intended purpose. | recommendation |
| EV-59 | "Medical device stand-alone software including apps (including IVDMDs)" v1.10f. UK MHRA, page updated 1 Jul 2023. [Landing page](https://www.gov.uk/government/publications/medical-devices-software-applications-apps), [PDF](https://assets.publishing.service.gov.uk/media/64a7d22d7a4c230013bba33c/Medical_device_stand-alone_software_including_apps__including_IVDMDs_.pdf) | Fitness monitoring is not usually a medical purpose. Lifestyle choices or referral advice ("see your GP") is unlikely device function. Influence on actual treatment is likely device function. | Newer version possible. Appendices not read. | recommendation |
| EV-60 | "Health Products Compliance Guidance." US FTC. [Page](https://www.ftc.gov/business-guidance/resources/health-products-compliance-guidance) | Health claims need competent and reliable scientific evidence, generally RCTs. Applies to health apps. | Issue date not shown (believed Dec 2022). | recommendation, date unresolved |
| EV-61 | FTC Health Breach Notification Rule, 16 CFR Part 318, final rule 89 FR 47028, 30 May 2024, effective 29 Jul 2024. [Rule page](https://www.ftc.gov/legal-library/browse/rules/health-breach-notification-rule), [GovInfo text](https://www.govinfo.gov/content/pkg/FR-2024-05-30/html/2024-10855.htm) | Fitness apps can be health care services and vendors of personal health records. Unauthorized disclosure counts as a breach. Notice within 60 days. | Enforcement practice not covered. | evidence (legal text) |
| EV-62 | "Health App Use Scenarios & HIPAA." HHS OCR, Feb 2016. [PDF, archive snapshot](https://www.hhs.gov/sites/default/files/ocr-health-app-developer-scenarios-2-2016.pdf) | An app that serves consumers directly, not on behalf of a covered entity, is not likely subject to HIPAA. | Changes for employer, provider, or payer distribution. Archive snapshot. | recommendation |
| EV-63 | Washington My Health My Data Act, RCW 19.373. [RCW 19.373.010](https://app.leg.wa.gov/RCW/default.aspx?cite=19.373.010), [AG overview](https://www.atg.wa.gov/protecting-washingtonians-personal-health-data-and-privacy) | Consumer health data includes symptoms, bodily functions, and inferences from algorithms. Consent to collect and share. Private right of action per the AG. | Operational details from the AG summary. | evidence (legal text) |
| EV-64 | Nevada consumer health data law, NRS 603A.400-603A.550 (SB 370, 2023). [NRS 603A](https://www.leg.state.nv.us/NRS/NRS-603A.html) | Similar definition of consumer health data. Consent and deletion duties. | Effective date and enforcement from secondary sources. | evidence (legal text) |
| EV-65 | Connecticut Data Privacy Act, CGS 42-515 and 42-526. [Chapter 743jj](https://www.cga.ct.gov/current/pub/chap_743jj.htm) | Consumer health data definition. No sale without consent. Geofencing ban near health facilities. | Applies to products targeted to Connecticut residents. | evidence (legal text) |
| EV-66 | New York Health Information Privacy Act (S929). No primary text verified. | Secondary reports: veto on 19 Dec 2025, revised bill in 2026. | nysenate.gov returned 403. | unresolved |

## 5. Synthesis

Everything in this section is synthesis. It is interpretation of the register, not new evidence. Each claim cites EV ids. The text marks each number that has no direct source as an assumption.

### 5.1 Resistance-training guidance for healthy adults (synthesis)

- ACSM 2026 is the current position stand for healthy adults. It replaces the 2009 progression-model stand (EV-1, EV-2).
- Its central message: any resistance program beats none, and adherence and individual fit matter more than exact prescription values (EV-1).
- Public-health guidance asks for muscle-strengthening work of moderate or greater effort on 2 or more days each week (EV-3, EV-4, EV-5).
- Four regions cover the major muscle groups: upper push, upper pull, lower push, lower pull (EV-1). A machine-only library (D-45) can cover all four.
- Machines and free weights give similar results (EV-1). The machine-only scope of D-45 does not reduce the expected outcome.
- Most emergency-department injuries in weight rooms come from weights that drop on the person, and most involve free weights (EV-52). Fixed-path machines (D-45) remove much of that mechanism.
- Injury rates in weight-training sports are low against team sports. The shoulder, low back, knee, elbow, and wrist carry most injuries (EV-51).
- For an intermediate lifter, autoregulation by RIR gives results similar to percentage-based loads (EV-39). An RIR log (D-57) is therefore a sound adaptation input.

### 5.2 Prescription variables (synthesis)

| Variable | What the evidence says | Candidate policy value for Workout App | Basis and class |
|---|---|---|---|
| Frequency | 2+ days/week for all major groups. Volume-equated frequency has little effect on hypertrophy but helps strength. | 2-3 sessions/week. Each of the four regions at least 2 times/week. | EV-1, EV-3, EV-16, EV-17, EV-19. Recommendation. |
| Weekly volume | Dose-response to 10+ sets/muscle/week, with diminishing returns. Indirect sets count as 0.5. | After a long break, start at about 6-8 direct sets per region per week. Build toward 10+ over 4-8 weeks. | EV-1, EV-18, EV-19. The ramp is an assumption. |
| Sets per exercise | 1 set works, 2-3 sets work better. | 2 sets in the first 1-2 weeks back, then 2-3. | EV-1, EV-3. Ramp is an assumption. |
| Rep range | Hypertrophy similar across loads at high effort. Strength favors heavier loads. RIR accuracy falls in sets over 12. | Default 8-12. Range limits 6-15. No default set above 15 reps. | EV-20, EV-21, EV-27. Recommendation. |
| Effort | Failure not required. 2-3 RIR is sufficient. Exact RIR targets cannot be quantified. | 1-3 RIR (D-37). RIR 3 in the first sessions after a break. | EV-1, EV-23 to EV-26. D-37 is an owner decision. |
| Rest | Over 60 s gives a small benefit, none beyond 90 s. Trained people gain strength with over 2 min. | Rest timer default 90-120 s for machines, 2-3 min for leg press and other large multi-joint machines (D-59). | EV-35, EV-36, EV-37. Recommendation. |
| Progression | Load step of 2-10% after the target is exceeded by 1-2 reps. Rep progression works as well as load progression. | Double progression inside a rep range. Rep progression when the rounded 5 lb step is too large (section 5.8). | EV-2, EV-38. Step limits are assumptions. |
| Order | Large before small, multi-joint before single-joint. Exercises early in a session gain more strength. | Multi-joint machines first. | EV-1, EV-2. Recommendation. |
| Recovery and deload | Deloads reduce volume and keep or reduce intensity. Reactive deloads follow performance and readiness. Full cessation costs some strength. | No planned block (D-43). Reactive deload: sets down 30-50%, RIR 3, same frequency. | EV-40, EV-41, EV-42. Triggers are assumptions. |

Adherence comes before optimization (EV-1). The first goal of the policy is a plan that the owner can finish safely and repeat.

### 5.3 Failure: evidence and limits (synthesis)

- Failure is not necessary for strength or hypertrophy (EV-1, EV-23, EV-24, EV-25).
- Strength gains are similar across a wide RIR range (EV-26).
- Hypertrophy can rise a little as sets end closer to failure (EV-26, low certainty). Trained people show a small signal (EV-23). The relation can be non-linear (EV-24).
- Failure adds acute fatigue and slows recovery by 24-48 h (EV-25, EV-33).
- ACSM flags form and vascular risk when some groups lift to fatigue (EV-1).
- Unaccustomed high effort and high volume carry a rare hazard signal for exertional rhabdomyolysis (EV-34).
- People underpredict the reps they have left by about 1 on average, with wide spread (EV-27, EV-28). Accuracy falls in sets over 12 (EV-27).
- Experience can improve RIR accuracy (EV-28, EV-29, EV-30), but one small study found no effect of experience (EV-31).

Consequences for the policy under D-37:

- Targets stay at 1 to 3 RIR (D-37).
- The policy never prescribes a set to failure in the first sessions after a break (D-37). No decision defines "first sessions", so section 7 lists it.
- A logged RIR of 0 counts as a failure event (assumption). The policy counts these events and holds load progression when they recur.
- Machines remove the drop hazard of free weights (EV-52). They do not remove the form and vascular concerns of EV-1.

### 5.4 Screening, warning signs, and pain (synthesis)

What the evidence recommends:

- The ACSM model screens three items: current activity, signs or symptoms or known disease, and desired intensity (EV-7, EV-6).
- The EIM form lists cardiac and vascular symptoms. Any positive answer means "STOP" and medical clearance before exercise (EV-8).
- The PAR-Q+ asks seven questions and adds delay rules for temporary illness and pregnancy (EV-9, EV-10).
- WHO says clearance is generally unnecessary for light or moderate activity without contraindications. It says new symptoms need a provider (EV-5).
- Vigorous activity in unfit people can acutely raise cardiac risk in susceptible people (EV-53). Cardiac events often follow warning signs (EV-7).

| Sign group | Examples | What the evidence recommends | Workout App behavior |
|---|---|---|---|
| Cardiorespiratory or neurological | Chest pain or pressure, unusual breathlessness, dizziness or fainting, palpitations, sudden severe headache, confusion | Stop exercise, seek medical clearance before a return (EV-8, EV-5, EV-53). Full GETP12 list not verified (EV-6). | Warn. The user can continue after a confirmation (D-40). No emergency advice (D-36). |
| Musculoskeletal | Sharp or stabbing pain, joint pain, pain that builds through a set, pop or snap, numbness, tingling, instability | Stop the movement. Pain-tolerant loading belongs to clinical care (EV-54, EV-55). | Warn and continue (D-40). The policy holds load progression for that exercise (section 5.8). |
| Illness | Cold, fever, recent health change | Delay the increase in activity (EV-9). | Not collected (D-34). |

Pain handling inside D-36 and D-40:

- Muscle burn and fatigue during a set are expected effort sensations. Diffuse soreness 24-72 h after a session is common after a return. These signs are not pain reports (assumption).
- The app does not diagnose pain and does not prescribe rehabilitation (D-36).
- A pain report (D-57) is a progression signal (D-64). The policy never increases load on the exercise that the report names.
- Referral advice such as "see a physician or physical therapist" stays inside wellness guidance (EV-59).

### 5.5 Gap against owner decisions (synthesis)

This subsection records what the evidence recommends that the owner decisions do not adopt. It records each difference as an accepted risk. It does not argue for a change.

| Evidence recommends | Source | Owner decision | Accepted risk |
|---|---|---|---|
| A readiness screen before a program: symptoms, known cardiovascular, metabolic, or renal disease, current activity, intended intensity, temporary illness. | EV-7, EV-8, EV-9, EV-5 | D-34: injuries and experience only, no readiness questions. | The app does not detect a known condition or symptom that the evidence routes to medical clearance. |
| A screen answer such as chest pain or a doctor's order for supervised activity stops automatic plans until clearance. | EV-8, EV-9 | D-35: a screen answer never stops plan generation. The app warns and continues with a plan that avoids the injured area. | The app generates plans for a user whom the evidence routes to a professional first. |
| Stop rules for cardiac and neurological red flags during exercise: stop the session, seek clearance before a return. | EV-8, EV-5, EV-53, EV-7 | D-40: warn, then continue after a confirmation, for every symptom. The owner declined a split rule for cardiac symptoms. | The user can continue a session after a cardiac warning sign. |
| Qualified human review of policy parameters, screening and stop rules, pain boundaries, and claims: strength coach, sports-medicine physician, physical therapist, counsel. A supervised pilot. | EV-1 (RIR not quantifiable), EV-6 (lists not verified), EV-54, EV-55, EV-56 | D-39: no qualified human review. | Assumed thresholds in sections 5.7 and 5.8 ship without expert sign-off. |

The core user profile adds one more difference. The evidence recommends "start low and go slow" for inactive people (EV-3, EV-53). D-32 treats the core user as intermediate from the first day with his own load estimates. Section 5.7 describes how the policy bounds those estimates inside D-32.

### 5.6 Effects of age, long breaks, injury, and other factors (synthesis)

Age:

- Resistance work is safe at all adult ages (EV-1). D-26 limits scope to adults 18 and older, and youth guidance assumes supervision (EV-15).
- Older adults lose more on a break (EV-43, EV-44). ACSM flags near-failure work for older people (EV-1). PAR-Q+ asks people aged 45 and older to consult a professional before vigorous effort (EV-9).
- Balance and multicomponent work help older adults (EV-5, EV-14). NSCA dosing is unresolved (EV-13).
- The core user is 32 (D-31), so no age rule applies now. The onboarding flow collects age (D-41), so the policy can apply age rules later.

Long breaks and detraining (important for the core user):

- Cessation reduces all muscular performance, and the loss grows with duration (EV-43).
- Muscle size that a program builds falls over months, with clear loss at 31-52 weeks (EV-44). After 20 weeks off, size returned to baseline (EV-47).
- Strength stays partly above the untrained level after a break, from motor learning (EV-47). A return to regular sessions recovers gains fast (EV-46).
- Benefits last longer after long programs (EV-45).
- Vigorous effort in an unfit person carries the highest acute cardiac risk (EV-53). Unaccustomed high volume and effort carry a rhabdomyolysis signal (EV-34).
- Consequence: the core user can relearn machine technique fast, but his load estimates from years ago can overstate his current capacity. Section 5.7 bounds them.

Injury:

- Onboarding collects injuries (D-34). The app warns and continues with a conservative plan that avoids the injured area (D-35).
- The user can exclude exercises (D-48). Pain-guided loading of an injury is clinical care (EV-54, EV-55) and stays outside D-36.

Disability, pregnancy, and chronic conditions:

- The scope decisions define the user as an experienced adult (D-26, D-30, D-33). No decision adds support for disability, pregnancy, or chronic conditions. The app does not screen for them (D-34).
- The evidence routes these groups to professional input: pregnancy (EV-11, EV-12, EV-5), chronic conditions (EV-3, EV-8, EV-9), and disability (EV-3, EV-5).
- The single user is a man (D-31, D-67), so pregnancy guidance has no current use.

### 5.7 Safe starting-load estimation (synthesis)

Evidence:

- Reps at a fixed %1RM vary widely between people (EV-22). A table value can mislead one person (EV-50).
- 1RM equations lose accuracy above 10 reps and show large errors on the leg press (EV-48, EV-49).
- RIR reports carry an error of about 1 rep on average, with wide spread (EV-27, EV-28).
- The evidence therefore supports load calibration by reps and reported effort, not by %1RM or a max test.

How the policy bounds the user estimates of D-32 (assumption unless a source is named):

1. The onboarding flow collects a load estimate for each machine (D-41). The policy treats each estimate as provisional.
2. The policy rejects an estimate outside the available weights of the machine (D-54).
3. After a break of more than 3 months, the first set uses at most 70-80% of the estimate.
4. The first exposure to each machine is one calibration set at the target reps.
5. The app cues a stop at 3 or more reps in reserve (D-37). The raw synthesis suggested 3-4 RIR.
6. The user logs reps and RIR (D-57). The policy adjusts with the table below.
7. The policy allows at most 3 calibration adjustments per machine per session. Calibration never becomes a max test.
8. The load stays "calibrating" for 1-2 more sessions.
9. The app never shows an estimated 1RM as a fact (EV-48).

| Calibration set result | Next set | Note under D-65 |
|---|---|---|
| RIR 6 or more, controlled, no pain | Up two 5 lb steps | At 25 lb, +10 lb is +40%. The policy checks the RIR after each step. |
| RIR 5 | Up one 5 lb step | At 25 lb, +5 lb is +20%. |
| RIR 3-4 | Keep the load | On target. |
| RIR 2 or less, form trouble, or pain | Down one 5 lb step | A pain report also triggers the D-40 warning. |

The rounding rule of D-65 applies to every start value. The policy checks the rounded value, not the computed value. When the rounded value exceeds a bound, the calibration cue and the RIR report catch the difference in the first set.

### 5.8 Adaptation rules (synthesis)

Base: double progression inside a rep range (EV-2, EV-38). Progression signals: reps, load, RIR, pain, skipped work, and gaps in the history (D-64). Each change carries a concise reason that names the logged evidence (D-68). An override keeps the recommendation, the override, and the reason as separate records (D-69).

Interaction with the 5 lb rounding of D-65:

- The evidence supports load steps of 2-10% (EV-2).
- Under D-65, the smallest step is 5 lb. At 25 lb one step is +20%, at 50 lb +10%, at 100 lb +5%.
- Nearest rounding can also move a computed load up by as much as 2.5 lb. An 8% step on 40 lb gives 43.2 lb, which rounds to 45 lb (+12.5%).
- So rounding up can exceed a target that the policy validated before the rounding. This result is an accepted consequence of D-65.
- The policy therefore validates the rounded load. When the rounded jump exceeds the policy limit (candidate 10%, assumption), the policy keeps the load and progresses reps or adds a set (EV-38).
- D-65 states no rule for a value halfway between two steps, such as 22.5 lb. Section 7 lists it.

Scenario A: prescribed 3 x 12 at 25 lb, achieved 12, 12, 5.

Evidence: progression needs the target met or exceeded (EV-2). A drop of 7 reps on set 3 suggests a load near the 12-13 RM, short rest (EV-36, EV-37), pain, or an interruption.

1. The policy does not increase the load.
2. A pain report on this exercise applies the pain rule first.
3. An early end or an interruption repeats the same prescription next session.
4. Rest under 90 s before set 3 repeats 3 x 12 at 25 lb with a 90-120 s rest target.
5. RIR 0-1 on sets 1 and 2 keeps 25 lb and lowers the target to 3 x 8-10.
6. The policy prefers this rep change over 20 lb, because one step down is -20%.
7. With no cause found, the policy repeats 3 x 12 at 25 lb.
8. The same shortfall in 2 consecutive sessions reduces the load to 20 lb at 3 x 10-12.
9. The policy never adds sets in the same session to make up missed reps.
10. RIR 0 on set 3 counts as a failure event (D-37).

Example reason text (D-68): "Set 3 ended 7 reps short. The load stays at 25 lb." The thresholds are assumptions. The "no increase" rule rests on EV-2.

Scenario B: all prescribed reps completed with RIR 3 or more on every set.

Evidence: this effort sits at the easy end of the D-37 range. Double progression adds load when the person exceeds the target (EV-2). Rep progression works as well as load progression (EV-38).

1. Reps below the top of the range: add 1-2 reps per set at the same load.
2. Reps at the top of the range: compute the next load, then round to the nearest 5 lb (D-65).
3. A rounded jump of 10% or less applies. Reps reset to the bottom of the range.
4. A rounded jump above 10% keeps the load. Reps rise toward 15, or the plan adds one set.
5. At 3 x 15 with RIR 3 or more, the policy adds 5 lb and resets reps to 8.
6. After calibration, the policy allows at most one 5 lb step per machine per session.
7. In the first sessions after a break, the policy uses rep progression only.

Worked example at RIR 3+: 3 x 12 at 25 lb goes to 3 x 14, then 3 x 15. Next comes 3 x 8 at 30 lb. The 10% limit, the 15-rep ceiling, and the reset to 8 are assumptions. Reps at a given %1RM vary between people (EV-22), so the policy reads RIR again on the first set at 30 lb.

Missed reps in general (assumption):

- A total shortfall of 2 reps or less across sets holds the load.
- A larger shortfall runs the checks of scenario A.
- A decline on 2 or more exercises over 2 or more sessions starts a reactive deload for one week (EV-40, EV-41). Sets go down 30-50%, RIR goes to 3, and frequency stays the same.
- The policy avoids a full stop, because a full week off costs some strength (EV-42).

Pain report:

- The app warns, and the user can continue after a confirmation (D-40).
- The policy records the report (D-57) and does not increase load on that exercise next session.
- The policy holds progression for that exercise until 2 sessions pass with no pain report (assumption).
- A pain report on the same exercise in 2 consecutive sessions adds referral text to the warning (EV-59). The text holds no diagnosis (D-36).
- The user can skip the exercise (D-47, D-63) or exclude it (D-48).

Interrupted session:

- "Finish now" skips the remaining exercises, and the session records as ended early (D-63).
- The policy counts the completed sets. When the user completed under 50-70% of the prescribed sets, the policy does not use the session for progression (assumption).
- An interruption after a symptom warning (D-40) holds all loads next session at RIR 3 (assumption).

Missed sessions and long breaks (D-66):

| Gap since the last session | Policy action (percentages are assumptions) | Evidence for direction |
|---|---|---|
| Under 2 weeks | Resume the last loads. RIR 3 for one session. | EV-46 |
| 2-4 weeks | About -10% load, about -30% sets, recalibrate if RIR is off. | EV-43, EV-46 |
| 1-3 months | About -20% load, reduced volume, 1-2 week ramp. | EV-43, EV-44 |
| Over 3 months | Restart calibration from the user estimate (section 5.7). | EV-43, EV-44, EV-47, EV-53 |

Under D-65, a -10% change on 25 lb gives 22.5 lb, a halfway value. The policy needs a tie rule for it (section 7). The core user returns after several years, so his first sessions fall in the last row.

### 5.9 Substitution principles (synthesis)

- A busy or unavailable machine leads to a skip, with no substitute (D-47). Substitution is out of scope now.
- A user can exclude an exercise or movement, with an optional reason (D-48). Luna plans again, and the policy checks the result.
- When Luna plans again after an exclusion, the policy can check these principles in order:

1. Safety constraints first: excluded movements and injured areas (D-35, D-48).
2. Movement pattern: keep the same pattern where a machine exists.
3. Regional coverage: keep weekly sets for each of the four regions (EV-1).
4. Load profile: a similar rep range, with a load that the 5 lb rounding can progress.
5. Available equipment: only machines that the user confirmed (D-49).

- Load does not transfer across machines. Every new machine starts with a calibration flag (section 5.7).
- Machines and free weights give similar outcomes (EV-1), so a machine change does not have to cost results.

### 5.10 Fitness versus medical device boundary (synthesis)

- FDA lists physical fitness claims as general wellness: track exercise, improve fitness, strength, endurance, or muscle size (EV-56).
- FDA names "treat muscle atrophy" and restoration of function impaired by disease as non-wellness examples (EV-56).
- The CDS exclusion needs a health care professional user. A consumer app cannot use it (EV-57).
- The EU and UK treat fitness apps as outside medical device software, with referral advice allowed (EV-58, EV-59).
- Health claims need competent and reliable scientific evidence (EV-60).
- All text that the product shows counts toward its intended use, and LLM output is part of that text (EV-56).
- D-36 limits the product to fitness guidance. D-81 removes public marketing pages, so Luna output is the main claims surface.
- No source in the register addresses software that one person builds for his own use (D-67). The boundary still guides the claims policy.

Claims that Luna must never make:

| Category | Example of a forbidden claim | Source |
|---|---|---|
| Diagnosis | "This sounds like tendinitis." | D-36, EV-56, EV-59 |
| Treatment or rehabilitation | "This plan rehabs your knee." "Treat back pain." | D-36, EV-56 |
| Disease management | "Manage your hypertension or diabetes with this plan." | EV-56, EV-57 |
| Named clinical conditions | "Reverse muscle atrophy or sarcopenia." | EV-56 |
| Clinical-looking physiologic values | A readiness score shown as a clinical value | EV-56 |
| Emergency advice | Instructions for a medical emergency | D-36 |
| Unsupported proof | "Clinically proven." "Physician-designed." | EV-60 |
| Safety guarantee | "This load is safe for you." | EV-1 (group-level evidence only) |

Allowed forms: fitness, strength, and adherence statements (EV-56). Generally accepted lifestyle statements in the second FDA category, such as "regular physical activity can help reduce the risk of high blood pressure" (EV-56). Referral advice (EV-59).

### 5.11 Privacy law notes (synthesis)

| Law | What it says | Relevance under D-67 and D-79 |
|---|---|---|
| HIPAA (EV-62) | A direct-to-consumer app is not likely subject to HIPAA. | Likely not applicable. |
| FTC Health Breach Notification Rule (EV-61) | Fitness apps can be vendors of personal health records. Unauthorized disclosure counts as a breach. | Low. The owner is the only consumer and also the operator. |
| Washington MHMDA (EV-63) | Symptoms, pain, and inferred health status are consumer health data. Consent to collect and share. | Low. The only consumer is the owner. No record gives the state of residence. |
| Nevada NRS 603A (EV-64), Connecticut CGS 42-515 (EV-65) | Similar consumer health data duties. | Low for the same reason. |
| New York HIPA (EV-66) | Status unresolved. | Low. |

D-67 limits the app to the owner, D-78 removes user data controls, and D-79 sets the minimum general consumer posture. These decisions reduce the relevance of the privacy laws. D-80 already keeps workout text and health details out of telemetry. The relevance returns if a later decision changes D-67.

### 5.12 Versioned evidence and policy model (synthesis)

Goal: every plan and every change is explainable, reproducible, and traceable to rules and evidence (D-23, D-68).

- Evidence registry: `EV-<n>` id, citation, URL or DOI, access date, class, limits, and a content hash of the extracted claim. Supersession links, for example EV-2 superseded by EV-1.
- Policy rule: rule id, rule version, intent, deterministic logic, evidence ids, assumption flags, safety tier, approver, and effective date. Under D-39 the approver is the owner.
- Policy bundle: a semver version of the rules, threshold tables, exercise library, and claims blocklist, with a changelog.
- Decision log per plan and per change: policy bundle version, rule ids fired and their outputs, and an input snapshot of the logged sets.
- The decision log also holds the model id (`gpt-6-luna`), the effort level (D-24), the prompt template hash, the Luna proposal, and the final output.
- The log keeps each load before and after the rounding of D-65. It also records rules fallback use when Luna fails (D-23).
- Separation: Luna proposes (D-22). The policy decides every set, load, and change (D-23). Luna never overrides a policy result.
- Change control: a new source starts an impact check on the rules that cite it. A new bundle version follows.

Golden tests:

| Test | Input | Expected policy output |
|---|---|---|
| Scenario A | 3 x 12 at 25 lb, logged 12, 12, 5 | No load increase. Branch by pain, interruption, rest, and RIR as section 5.8 lists. |
| Scenario B at 25 lb | 3 x 12 at 25 lb, all reps, RIR 3+ | Load stays 25 lb. Reps rise, because 30 lb is +20%. |
| Scenario B at 100 lb | 3 x 12 at 100 lb, all reps, RIR 3+ | Load rises to 105 lb (+5%). Reps reset to 8. |
| Rounding up | Computed 43.2 lb from 40 lb | Rounded 45 lb is +12.5%, so the policy keeps 40 lb and adds reps. |
| Pain report | Pain logged on one exercise | D-40 warning. No load increase on that exercise next session. |
| Long break | Last session 4 years ago | Calibration from the user estimate at 70-80%, RIR 3 cue, rep progression only. |
| Failure guard | First session after a break | No prescribed set below RIR 3. |
| Claims filter | Luna text with "treat" or "rehab" | Blocked or rewritten before display. |

Property tests check four invariants. No load increase follows a shortfall. No rounded jump exceeds the policy limit. No failure target appears in the first sessions after a break. No output holds a blocked claim.

### 5.13 Qualified human review (synthesis)

- The evidence and the raw synthesis recommend qualified review before release (EV-1, EV-6).
- Programming parameters: a certified strength coach, because no source quantifies exact RIR targets (EV-1).
- Screening and stop rules: a sports-medicine physician, because the termination list is not verified (EV-6).
- Pain boundaries: a physical therapist (EV-54, EV-55).
- Claims and privacy: counsel (EV-56, EV-61).
- A supervised pilot with adverse-event records, because no study tests this delivery model.
- Owner decision: no qualified human review is necessary before broader release (D-39). This document records the decision as an accepted risk in section 5.5.

## 6. Recommendations not yet owner decisions

Every row in this table is a recommendation. No row is an owner decision.

| Id | Recommendation | Basis | Related decisions |
|---|---|---|---|
| REC-1 | Recommendation: validate the rounded load. Limit a rounded jump to 10%, and use rep progression when the jump is larger. | EV-2, EV-38, assumption | D-23, D-65 |
| REC-2 | Recommendation: add a tie rule for halfway values. Round down on an increase and on a return after a break. | Assumption | D-65 |
| REC-3 | Recommendation: allow a downward rounding exception when nearest rounding pushes a start load above a safety bound. This change amends D-65. | Assumption | D-65, D-32 |
| REC-4 | Recommendation: define "first sessions after a break" as the first 3 sessions or 14 days, whichever is longer. Use RIR 3 and rep progression only. | EV-43, EV-53, assumption | D-37, D-66 |
| REC-5 | Recommendation: adopt the calibration protocol and table of section 5.7, including 70-80% of the estimate after a break over 3 months. | EV-22, EV-27, EV-48, assumption | D-32, D-41 |
| REC-6 | Recommendation: adopt the scenario A and scenario B rules of section 5.8 as policy rules with golden tests. | EV-2, EV-38 | D-23, D-64 |
| REC-7 | Recommendation: adopt reactive deload triggers inside continuous adaptation. | EV-40, EV-41, EV-42 | D-43 |
| REC-8 | Recommendation: adopt the long-break table of section 5.8. | EV-43 to EV-47 | D-66 |
| REC-9 | Recommendation: hold load after a pain report until 2 sessions pass with no report. Add referral text after 2 consecutive reports. | EV-54, EV-55, EV-59 | D-36, D-40, D-57 |
| REC-10 | Recommendation: define the warning text so that it stays inside D-36 and names the symptom plainly. | EV-8, EV-59 | D-36, D-40 |
| REC-11 | Recommendation: add a versioned blocked-claims filter on all Luna text. | EV-56, EV-60 | D-22, D-36 |
| REC-12 | Recommendation: log model id, effort, prompt template hash, policy bundle version, rules fired, and loads before and after rounding. | Section 5.12 | D-23, D-24, D-68 |
| REC-13 | Recommendation: start weekly volume after a long break at about 6-8 direct sets per region, and build toward 10+ over 4-8 weeks. | EV-1, EV-18, EV-19, EV-34 | D-32, D-43 |
| REC-14 | Recommendation: set rest timer defaults of 90-120 s for machines and 2-3 min for large multi-joint machines. | EV-36, EV-37 | D-59 |
| REC-15 | Recommendation: count a logged RIR of 0 as a failure event, and hold progression when such events recur. | EV-23, EV-33, EV-34 | D-37, D-57 |
| REC-16 | Recommendation: check the privacy notes of section 5.11 again if a later decision changes D-67. | EV-61 to EV-65 | D-67, D-79 |

## 7. Unresolved items

- D-65 states no tie rule for a value halfway between two 5 lb steps, such as 22.5 lb.
- D-65 rounds to 5 lb, but a machine can have other steps (D-54). The policy has no rule for a rounded load that the machine does not have.
- No decision defines "the first sessions after a break" of D-37.
- The raw synthesis suggests calibration at 3-4 RIR, but D-37 sets targets at 1-3 RIR. The owner has no decision on this difference.
- D-36 excludes emergency advice. The boundary between a symptom warning (D-40) and emergency advice has no definition.
- The device status of software that one person builds for his own use is not in the register (D-67).
- No record gives the state of residence of the owner, so the relevance of the Washington, Nevada, and Connecticut laws is unresolved.
- ACSM GETP12 termination criteria and symptom tables: not accessed (EV-6).
- Riebe 2015 full text: paywalled (EV-7). The EIM form verified the model (EV-8).
- NSCA older-adult dosing: abstract only (EV-13).
- ACOG contraindication tables: images, not transcribed (EV-11). Canadian pregnancy guideline lists: abstract only (EV-12).
- Pelland 2026 diminishing-return points: not extracted (EV-19).
- Bosquet 2013 duration-specific effects: abstract truncated (EV-43).
- LeSuer 1997 1RM equations: not located in Europe PMC.
- PAR-Q+ licensing for app use: not found (EV-9). D-34 makes this item moot for now.
- FTC Health Products Compliance Guidance issue date: not shown (EV-60).
- Nevada effective date and enforcement: from secondary sources only (EV-64).
- New York HIPA status: unresolved (EV-66).
- Other state laws on sensitive data, EU GDPR Article 9, the EU AI Act, and UK GDPR: not researched. D-28 limits the market to US English.
- App-store health policies: not researched. D-17 removes app stores, so this item has no current use.
- Validation of photo-based machine recognition for safety use: no evidence found. D-49 requires user confirmation of every machine.
