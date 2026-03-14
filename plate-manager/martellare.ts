import { log } from "./log";

const LYRICS = `
[00:00]^^MARTELLARE // SadTurs, KIID, BelloFigo
[00:01]^^A-a-amaiadaiafaiachia
[00:02]^^SadTurs, ahahah
[00:03]^^KIID, you ready?
[00:04]^^Tutti sanno che vogliamo martellare
[00:05]^^Yung Snapp made another hit
[00:07]^^Ehi, ehi
[00:11]^^Tatuaggi addosso, arrivo con dell'ice
[00:13]^^Driftando nella Porsche, non mi chiedere: "Come stai?"
[00:15]^^Vogliamo martellare, martellare
[00:17]^^Martellare, martellare, martellare, martellare, martellare
[00:20]^^Martellare, martellare
[00:22]^^La tua ragazza va giu come una foglia
[00:25]^^Dice che sono carino, nel club mi sta tutta addosso
[00:27]^^Mi fara martellare, martellare
[00:28]^^Martellare, martellare, martellare, martellare, martellare
[00:34]^^Lei mi chiede un ballo ma non so ballare
[00:37]^^Faccio andare giu le tipe, ti insegno come fare
[00:40]^^Non mi muovo, resto fermo, contanti in tasca
[00:43]^^Twerkava sui contanti, pensava fosse il mio zzoca
[00:45]^^Lei mi fa: "Insegnami a martellare
[00:47]^^Insegnami a martellare
[00:49]^^Insegnami a martellare
[00:51]^^Insegnami a martellare
[00:52]^^Insegnami a martellare"
[00:54]^^Prima assaggio in bagno, poi torno in pista a ballare sul terreno
[00:57]^^Vai martello, fai il martello
[01:00]^^Nuovo ballo
[01:02]^^Martello, vai martello
[01:04]^^Lascia che ti spieghi perche piaccio alla tua tipa
[01:08]^^Non ho la tartaruga, c'ho l'anaconda
[01:09]^^Tatuaggi addosso, arrivo con dell'ice
[01:12]^^Driftando nella Porsche, non mi chiedere: "Come stai?"
[01:15]^^Vogliamo Martellare, martellare, martellare, martellare, martellare
[01:18]^^Martellare, martellare, martellare, martellare, martellare
[01:19]^^La tua ragazza va giu come una foglia
[01:22]^^Dice che sono carino, nel club mi sta tutta addosso
[01:25]^^Mi fara martellare, martellare
[01:26]^^Martellare, martellare, martellare, martellare, martellare
[01:31]^^Io e i miei amici siamo tutti in canottiere
[01:34]^^Davanti al locale, non ci vogliono far entrare
[01:38]^^Fra', vedete di non farci arrabbiare
[01:41]^^Che abbiam le fighe bianche nel locale ad aspettare
[01:45]^^In prive solo fighe, niente maschi
[01:48]^^Ordiniamo le bottiglie, non guardiamo prezzi
[01:52]^^Da come mi twerka, oh mio Dio, ho perso i sensi
[01:53]^^Sono nero, tatuato quindi sai che sono sexy
[01:57]^^Vai martello, fai il martello
[02:00]^^Martello, vai martello
[02:02]^^Uah, uah, uah, scopo piu di un DJ
[02:07]^^Uah, uah, uah, scopo piu di un calciatore
[02:11]^^Tatuaggi addosso, arrivo con dell'ice
[02:16]^^Driftando nella Porsche, non mi chiedere: "Come stai?"
[02:20]^^Vogliamo Martellare, martellare, martellare, martellare, martellare
[02:27]^^Martellare, martellare, martellare, martellare, martellare
[02:31]^^Uah, uah, uah, scopo piu di un DJ
[02:34]^^Uah, uah, uah, scopo piu di un calciatore
`;

const PARSED_LYRICS = LYRICS.split("\n")
  .map((line) => {
    const match = line.match(/^\[(\d{2}):(\d{2})\]\^\^(.*)$/);
    if (!match) return null;
    // @ts-ignore
    const minutes = parseInt(match[1], 10);
    // @ts-ignore
    const seconds = parseInt(match[2], 10);
    const text = match[3];
    return {
      time: minutes * 60 + seconds,
      text,
    };
  })
  .filter((line): line is { time: number; text: string } => line !== null);

export async function stream_lyrics() {
  const { readable, writable } = new TransformStream();
  const writer = writable.getWriter();

  (async () => {
    try {
      const startMs = Date.now();

      for (let i = 0; i < PARSED_LYRICS.length; i++) {
        const line = PARSED_LYRICS[i];
        if (!line) {
          await writer.write("\n");
          continue;
        }

        const targetLineStartMs = line.time * 1000;
        const elapsedMs = Date.now() - startMs;
        const waitUntilLineMs = targetLineStartMs - elapsedMs;
        if (waitUntilLineMs > 0) {
          await new Promise((resolve) => setTimeout(resolve, waitUntilLineMs));
        }

        const words = line.text.trim().split(/\s+/).filter(Boolean);
        if (words.length === 0) {
          await writer.write("\n");
          continue;
        }

        const nextLine = PARSED_LYRICS[i + 1];
        const availableWindowMs = nextLine
          ? Math.max(0, (nextLine.time - line.time) * 1000)
          : 0;
        const perWordDelayMs =
          words.length > 1 ? availableWindowMs / words.length : 0;

        for (let w = 0; w < words.length; w++) {
          const chunk = w < words.length - 1 ? `${words[w]} ` : words[w];

          if (!chunk) continue;

          for (let c = 0; c < chunk.length; c++) {
            await writer.write(chunk[c]);
            await new Promise((resolve) => setTimeout(resolve, 2));
          }

          if (w < words.length - 1 && perWordDelayMs > 0) {
            const targetWordMs = targetLineStartMs + perWordDelayMs * (w + 1);
            const nowElapsedMs = Date.now() - startMs;
            const waitMs = targetWordMs - nowElapsedMs;
            if (waitMs > 0) {
              await new Promise((resolve) => setTimeout(resolve, waitMs));
            }
          }
        }

        await writer.write("\n");
      }
    } catch (error) {
      log("Error streaming lyrics:", error);
    } finally {
      await writer.close();
    }
  })();

  return new Response(readable, {
    headers: {
      "Content-Type": "text/plain",
      "Transfer-Encoding": "chunked",
      "X-Content-Type-Options": "nosniff",
    },
  });
}
