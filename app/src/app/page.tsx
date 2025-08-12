"use client";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useUserService } from "@/hooks/services/useUserService";

const Page = () => {
  const userService = useUserService();
  const [data, setData] = useState<string>("");
  const handleGetUser = async () => {
    console.log("CLICKED");
    const user = await userService.createUser({
      name: "tony",
    });
    const userJSON = JSON.stringify(user);
    setData(userJSON);
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <Button onClick={() => handleGetUser()}>Click Me</Button>
      <div>
        <strong>{data}</strong>
      </div>
    </div>
  );
};

export default Page;
