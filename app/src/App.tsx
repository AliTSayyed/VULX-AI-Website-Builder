import { useState } from "react";
import "./Styles/App.css";
import { Button } from "./components/ui/button";
import { useUserService } from "./hooks/services/useUserService";

function App() {
  const userService = useUserService();
  const [data, setData] = useState<string>("Click Me");
  const handleGetUser = async () => {
    console.log("CLICKED");
    const user = await userService.getUser({ id: "123" });
    const userJSON = JSON.stringify(user);
    setData(userJSON);
  };

  const click = function () {
    console.log("CLICKED");
  };

  return <Button onClick={() => handleGetUser()}>{data}</Button>;
}

export default App;
