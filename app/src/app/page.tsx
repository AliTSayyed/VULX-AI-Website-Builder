"use client";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { useUserService } from "@/hooks/services/useUserService";
import { User, UserSchema } from "@apiv1/user_service_pb";
import { create } from "@bufbuild/protobuf";

const Page = () => {
  const userService = useUserService();
  const [createUser, setCreateUser] = useState<User>(create(UserSchema));
  const [retrievedUser, setRetrievedUser] = useState<User>(create(UserSchema));

  const handleCreateUser = async () => {
    console.log("Creating user button clicked");
    const response = await userService.createUser({
      name: "tony",
    });
    // Check if response or response.user is null/undefined
    if (!response || !response.user) {
      // Handle the case where there's no user in the response
      console.error("No user returned from createUser");
      return; // or set a default value
    }

    setCreateUser(response.user);
  };

  const handleGetUsers = async () => {
    console.log("Get all users button clicked");
    const response = await userService.getUser({
      id: "",
    });
    // Check if response or response.user is null/undefined
    if (!response || !response.user) {
      // Handle the case where there's no user in the response
      console.error("No user returned from getUser");
      return; // or set a default value
    }

    setRetrievedUser(response.user);
  };

  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <Button onClick={() => handleCreateUser()}>Create User</Button>
      <div>
        <strong>
          created {createUser.name} with an id of: {createUser.id}
        </strong>
      </div>
      <Button onClick={() => handleGetUsers()}>Get Users</Button>
      <strong>
        Retreived {retrievedUser.name} with an id of: {retrievedUser.id}
      </strong>
    </div>
  );
};

export default Page;
