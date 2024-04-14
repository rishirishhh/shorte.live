import { Button } from "@/components/ui/button";
import { Link2Icon } from "@radix-ui/react-icons";
import { HeadingFour, HeadingTwo, Paragraph } from "./typography";
import { useToast } from "@/components/ui/use-toast";

export default function URLCard({
  shortUrl,
  destination,
  expiry,
  onEditClick,
  onDeleteClick,
  onStatsClick,
}: {
  shortUrl: string;
  destination: string;
  expiry: Date;
  onEditClick: () => void;
  onDeleteClick: () => void;
  onStatsClick: () => void;
}) {
  const { toast } = useToast();
  const CopyToClipboard = (value: string, message?: string) => {
    navigator.clipboard.writeText(value);
    toast({
      title: message ?? "Successfully copied to your clipboard",
      duration: 2000,
    });
  };

  return (
    <div className="w-full max-w-5xl bg-primary-foreground p-8 sm:p-10 rounded-md flex justify-between items-start text-left flex-col md:flex-row text-sm sm:text-lg gap-4">
      {/* content */}
      <div className="flex justify-center items-start flex-col gap-4 w-full md:w-4/6 ">
        {/* heading */}
        <div className="">
          <HeadingTwo className="w-max text-xl hover:cursor-pointer sm:text-2xl">
            <button className="w-max" onClick={() => CopyToClipboard(shortUrl)}>
              {shortUrl}
            </button>
            <Link2Icon className="inline-block ml-2 w-5 h-5" />
          </HeadingTwo>
          <HeadingFour className="text-slate-500 text-base dark:text-slate-400 w-[35ch]">
            {destination}
          </HeadingFour>
        </div>

        {/* description */}
        <div>
          <Paragraph className="text-slate-500 text-base dark:text-slate-300">
            <b>Expiry</b>-{expiry.toLocaleString()}
          </Paragraph>
        </div>
      </div>

      {/* action */}
      <div className="flex justify-end items-center gap-4 w-full ">
        <Button
          onClick={onEditClick}
          variant="outline"
          className="font-bold md:w-1/6 text-sm sm:text-base"
        >
          Edit
        </Button>
        <Button
          onClick={onDeleteClick}
          variant="destructive"
          className="font-bold md:w-1/6 text-sm sm:text-base"
        >
          Delete
        </Button>
        <Button
          onClick={onStatsClick}
          variant="secondary"
          className="font-bold md:w-1/6 text-sm sm:text-base"
        >
          Stats
        </Button>
      </div>
    </div>
  );
}
