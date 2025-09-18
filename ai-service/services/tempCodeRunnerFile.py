class OpenAICodeAgentService:
    def __init__(self, openai_client: OpenAIClient, sandbox_service:SandboxService):
        self.llm = openai_client.get_client()
        code_agent_tools = sandbox_service.get_tools()
        self.parser = PydanticOutputParser(pydantic_object=CodeAgentResult)
        self.prompt = ChatPromptTemplate.from_template(
        template=f"""
        {NEXTJS_PROMPT}
        {self.parser.get_format_instructions()}
        """
        )
        code_agent = create_tool_calling_agent(
                llm=self.llm,
                prompt=self.prompt,
                tools=code_agent_tools 
            )
        self.agent = AgentExecutor(agent=code_agent, tools=code_agent_tools, verbose=True)
            

    async def process_code_request(self, sandbox_id:str, user_message:str):
        try: 
            contextual_input:str = f"Sandbox ID: {sandbox_id}\nTask: {user_message}"
            callback = CodeAgentCallBack()
            result = await self.agent.ainvoke({"input": contextual_input}, config={"callbacks":[callback]})
            logger.info(result)
            parsed_result = self.parser.parse(result["output"])
            agent_actions = callback.get_result()

            return CodeAgentData(
                summary=parsed_result.summary,
                commands=agent_actions.commands_executed,
                files=agent_actions.updated_files
            ) 
        except Exception as e:
            raise f"openai code agent failed to generate response. error: {str(e)}"