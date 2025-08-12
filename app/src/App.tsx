import { useState } from "react";
import "./Styles/App.css";
import { Button } from "./components/ui/button";
import { useUserService } from "./hooks/services/useUserService";

function App() {
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
    <>
      <Button onClick={() => handleGetUser()}>Click Me</Button>
      <div>
        <strong>{data}</strong>
      </div>
    </>
  );
}

export default App;
