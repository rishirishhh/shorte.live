import { forwardRef, useEffect, useState } from "react";
import fetch from "@/utils/axios";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { useMain } from "./main-provider";

const TimeSlots = {
  "1h": {
    label: "Last 1 hour",
    hours: 1,
  },
  "1d": {
    label: "Last 24 hours",
    hours: 24,
  },
  "1w": {
    label: "Last 1 week",
    hours: 168,
  },
  "1m": {
    label: "Last 1 month",
    hours: 720,
  },
  "1y": {
    label: "Last 1 year",
    hours: 8760,
  },
};

function StatsURL(
  {
    urlObj,
  }: {
    urlObj: any;
  },
  ref: any
) {
  const [stats, setStats] = useState<any>(null);
  const { userState } = useMain();
  const [timeSlot, setTimeSlot] = useState(TimeSlots["1d"]);

  useEffect(() => {
    if (userState.login && urlObj) {
      (async () => {
        let endTime = new Date().getTime();
        let startTime = new Date(
          endTime - timeSlot.hours * 60 * 60 * 1000
        ).getTime();

        endTime = Math.floor(endTime / 1000);
        startTime = Math.floor(startTime / 1000);

        const res = await fetch.get(`/url/${urlObj._id}/stats?start=${startTime}&end=${endTime}`, {
          withCredentials: true,
        });
        setStats(res.data.data);
      })();
    }
  }, [userState.login, urlObj?._id]);

  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button ref={ref} style={{ display: "none" }}></Button>
      </DialogTrigger>
      <DialogContent className="min-w-[70vw] h-max p-8">
        <DialogHeader>
          <DialogTitle className="text-2xl">
            Stats for your URL ( beta )
          </DialogTitle>
          <DialogDescription className="text-lg">
            Stats for {urlObj?.short}
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-wrap w-full gap-4">
          <div className="flex-grow w-[200px]">
            <h2 className="text-lg">Device Counts</h2>
            <ul className="p-1">
              {Object.keys(stats?.device_counts ?? {})?.map((statKey: any) => {
                if (statKey === "total") return null;
                return (
                  <li key={statKey}>
                    {statKey.slice(6)} - {stats?.device_counts[statKey]}
                  </li>
                );
              })}
            </ul>
          </div>
          <div className="flex-grow w-[200px]">
            <h2 className="text-lg">OS Counts</h2>
            <ul className="p-1">
              {Object.keys(stats?.os_counts ?? {})?.map((statKey: any) => {
                if (statKey === "total") return null;
                return (
                  <li key={statKey}>
                    {statKey.slice(2)} - {stats?.os_counts[statKey]}
                  </li>
                );
              })}
            </ul>
          </div>
          <div className="flex-grow w-[200px]">
            <h2 className="text-lg">Geo Counts</h2>
            <ul className="p-1">
              {Object.keys(stats?.geo_counts ?? {})?.map((statKey: any) => {
                if (statKey === "total") return null;
                return (
                  <li key={statKey}>
                    {statKey} - {stats?.geo_counts[statKey]}
                  </li>
                );
              })}
            </ul>
          </div>
          <div className="flex-grow w-[200px]">
            <h2 className="text-lg">Referrer Counts</h2>
            <ul className="p-1">
              {Object.keys(stats?.referrer_counts ?? {})?.map(
                (statKey: any) => {
                  if (statKey === "total") return null;
                  return (
                    <li key={statKey}>
                      {statKey} - {stats?.referrer_counts[statKey]}
                    </li>
                  );
                }
              )}
            </ul>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}

const StatsURLModal = forwardRef(StatsURL);

export default StatsURLModal;
